package servers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

// FleetNotificationTest describes Master's Telegram routing for a fleet test message.
type FleetNotificationTest struct {
	MasterName   string
	HubName      string
	HubIsLocal   bool
	ViaMaster    []string
	Local        []string
	Disabled     []string
	Unreachable  []string
	DeliveryHint string
}

// CollectFleetNotificationTest gathers hub + per-panel notification_mode for Master test sends.
// ok=false when this panel is not Master.
func (s *Service) CollectFleetNotificationTest(ctx context.Context) (FleetNotificationTest, bool, error) {
	settings, err := s.ensureSettings(ctx)
	if err != nil {
		return FleetNotificationTest{}, false, err
	}
	if settings.Mode != ModeMaster {
		return FleetNotificationTest{}, false, nil
	}

	out := FleetNotificationTest{
		MasterName: strings.TrimSpace(settings.NodeName),
		ViaMaster:  nil,
		Local:      nil,
		Disabled:   nil,
	}
	if out.MasterName == "" {
		out.MasterName = "Master"
	}

	hubName := out.MasterName
	hubLocal := true
	if settings.NotificationHubNodeID.Valid {
		hubID := uuid.UUID(settings.NotificationHubNodeID.Bytes)
		if local, err := s.q.GetLocalServersNode(ctx); err == nil && local.ID == hubID {
			hubName = local.Name
			hubLocal = true
		} else if node, err := s.q.GetServersNode(ctx, hubID); err == nil {
			hubName = node.Name
			hubLocal = node.ConnectionType == ConnLocal
		} else {
			hubName = hubID.String()
			hubLocal = false
		}
	}
	out.HubName = hubName
	out.HubIsLocal = hubLocal
	if hubLocal {
		out.DeliveryHint = "Master local Telegram"
	} else {
		out.DeliveryHint = "hub panel " + hubName
	}

	// Master itself always delivers through hub/local (not "via master").
	out.Local = append(out.Local, out.MasterName+" (Master)")

	nodes, err := s.q.ListServersNodes(ctx)
	if err != nil {
		return out, true, mapErr(err)
	}
	for _, node := range nodes {
		if node.DeletedAt.Valid {
			continue
		}
		if node.ConnectionType == ConnLocal {
			continue
		}
		if !IsBarnPanel(node.ConnectionType) {
			continue
		}
		mode, err := s.fetchRemoteNotificationMode(ctx, node.ID, node.BaseUrl)
		if err != nil {
			out.Unreachable = append(out.Unreachable, node.Name)
			continue
		}
		switch mode {
		case NotifyMaster:
			out.ViaMaster = append(out.ViaMaster, node.Name)
		case NotifyDisabled:
			out.Disabled = append(out.Disabled, node.Name)
		default:
			out.Local = append(out.Local, node.Name)
		}
	}
	return out, true, nil
}

func (s *Service) fetchRemoteNotificationMode(ctx context.Context, nodeID uuid.UUID, baseURL string) (string, error) {
	base := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if base == "" {
		return "", fmt.Errorf("no base url")
	}
	cred, err := s.q.GetOutboundCredential(ctx, nodeID)
	if err != nil || len(cred.EncryptedToken) == 0 {
		return "", fmt.Errorf("no credentials")
	}
	token, err := s.cipher.Decrypt(cred.EncryptedToken)
	if err != nil || strings.TrimSpace(token) == "" {
		return "", fmt.Errorf("decrypt credentials")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, base+"/api/servers/settings", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-Barn-Token", token)
	client := &http.Client{Timeout: 8 * time.Second, CheckRedirect: limitRedirects}
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if res.StatusCode >= 300 {
		return "", fmt.Errorf("status %d", res.StatusCode)
	}
	var body struct {
		NotificationMode string `json:"notification_mode"`
	}
	if err := json.Unmarshal(raw, &body); err != nil {
		return "", err
	}
	mode := strings.TrimSpace(body.NotificationMode)
	if mode == "" {
		mode = NotifyLocal
	}
	return mode, nil
}
