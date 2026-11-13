package servers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

// RelayTelegram implements notifications.HubRelay.
// On Master, when a remote Barn panel is selected as the notification hub,
// Telegram delivery is POSTed to that panel; otherwise the caller sends locally.
func (s *Service) RelayTelegram(ctx context.Context, text string) (bool, error) {
	settings, err := s.ensureSettings(ctx)
	if err != nil {
		return false, err
	}
	if settings.Mode != ModeMaster || !settings.NotificationHubNodeID.Valid {
		return false, nil
	}
	hubID := uuid.UUID(settings.NotificationHubNodeID.Bytes)
	local, err := s.q.GetLocalServersNode(ctx)
	if err == nil && local.ID == hubID {
		return false, nil
	}
	node, err := s.q.GetServersNode(ctx, hubID)
	if err != nil {
		return true, fmt.Errorf("notification hub node: %w", mapErr(err))
	}
	if node.DeletedAt.Valid {
		return true, fmt.Errorf("notification hub node was deleted")
	}
	if !IsBarnPanel(node.ConnectionType) && node.ConnectionType != ConnLocal {
		return true, fmt.Errorf("notification hub must be a Barn panel")
	}
	if node.ConnectionType == ConnLocal {
		return false, nil
	}
	base := strings.TrimRight(strings.TrimSpace(node.BaseUrl), "/")
	if base == "" {
		return true, fmt.Errorf("notification hub has no base URL")
	}
	cred, err := s.q.GetOutboundCredential(ctx, node.ID)
	if err != nil || len(cred.EncryptedToken) == 0 {
		return true, fmt.Errorf("missing credentials for notification hub")
	}
	token, err := s.cipher.Decrypt(cred.EncryptedToken)
	if err != nil || strings.TrimSpace(token) == "" {
		return true, fmt.Errorf("cannot decrypt hub credentials")
	}
	body, _ := json.Marshal(map[string]string{"text": text})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, base+"/api/notifications/relay", bytes.NewReader(body))
	if err != nil {
		return true, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-Barn-Token", token)
	client := &http.Client{Timeout: 20 * time.Second, CheckRedirect: limitRedirects}
	res, err := client.Do(req)
	if err != nil {
		return true, fmt.Errorf("hub telegram relay failed: %w", err)
	}
	defer res.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if res.StatusCode >= 300 {
		return true, fmt.Errorf("hub telegram relay: %s", strings.TrimSpace(string(raw)))
	}
	return true, nil
}
