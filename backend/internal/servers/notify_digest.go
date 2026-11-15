package servers

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

// CollectRemoteDigestTexts fetches each remote Barn panel's status digest text.
// Local Master is excluded — the caller appends its own digest last.
// ok=false when this panel is not Master.
func (s *Service) CollectRemoteDigestTexts(ctx context.Context) ([]string, bool, error) {
	settings, err := s.ensureSettings(ctx)
	if err != nil {
		return nil, false, err
	}
	if settings.Mode != ModeMaster {
		return nil, false, nil
	}

	nodes, err := s.q.ListServersNodes(ctx)
	if err != nil {
		return nil, true, mapErr(err)
	}

	out := make([]string, 0, len(nodes))
	for _, node := range nodes {
		if node.DeletedAt.Valid || node.ConnectionType == ConnLocal || !IsBarnPanel(node.ConnectionType) {
			continue
		}
		name := strings.TrimSpace(node.Name)
		if name == "" {
			name = node.ID.String()
		}
		text, err := s.fetchRemoteDigestText(ctx, node.ID, node.BaseUrl, name)
		if err != nil {
			out = append(out, fmt.Sprintf("<b>Barn — %s</b>\n❌ Не удалось получить статус", html.EscapeString(name)))
			continue
		}
		text = strings.TrimSpace(text)
		if text == "" {
			continue
		}
		out = append(out, text)
	}
	return out, true, nil
}

func (s *Service) fetchRemoteDigestText(ctx context.Context, nodeID uuid.UUID, baseURL, nodeName string) (string, error) {
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

	status, raw, err := s.nodeGET(ctx, base+"/api/notifications/digest", token)
	if err != nil {
		return "", err
	}
	if status == http.StatusOK {
		var body struct {
			Text string `json:"text"`
		}
		if err := json.Unmarshal(raw, &body); err != nil {
			return "", err
		}
		if strings.TrimSpace(body.Text) == "" {
			return "", fmt.Errorf("empty digest")
		}
		return body.Text, nil
	}
	// Older panels: build compact digest from sites/databases health.
	if status == http.StatusNotFound || status == http.StatusMethodNotAllowed {
		return s.buildRemoteDigestFallback(ctx, base, token, nodeName)
	}
	return "", fmt.Errorf("status %d: %s", status, strings.TrimSpace(string(raw)))
}

func (s *Service) buildRemoteDigestFallback(ctx context.Context, base, token, nodeName string) (string, error) {
	panelName := nodeName
	if status, raw, err := s.nodeGET(ctx, base+"/api/notifications/settings", token); err == nil && status == 200 {
		var settings struct {
			PanelName string `json:"panel_name"`
		}
		if json.Unmarshal(raw, &settings) == nil && strings.TrimSpace(settings.PanelName) != "" {
			panelName = strings.TrimSpace(settings.PanelName)
		}
	}

	names := map[string]string{}
	if status, raw, err := s.nodeGET(ctx, base+"/api/sites", token); err == nil && status == 200 {
		var sites []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		}
		if json.Unmarshal(raw, &sites) == nil {
			for _, site := range sites {
				names[site.ID] = site.Name
			}
		}
	}

	var issues strings.Builder
	if status, raw, err := s.nodeGET(ctx, base+"/api/sites/health", token); err == nil && status == 200 {
		var rows []struct {
			SiteID  string `json:"site_id"`
			Overall string `json:"overall"`
			Message string `json:"message"`
		}
		if json.Unmarshal(raw, &rows) == nil {
			for _, row := range rows {
				if row.Overall == "healthy" {
					continue
				}
				name := names[row.SiteID]
				if name == "" {
					name = row.SiteID
				}
				fmt.Fprintf(&issues, "%s Сайт: %s — %s", remoteStatusEmoji(row.Overall), html.EscapeString(name), remoteOverallLabel(row.Overall))
				if msg := strings.TrimSpace(row.Message); msg != "" {
					fmt.Fprintf(&issues, ": %s", html.EscapeString(msg))
				}
				issues.WriteByte('\n')
			}
		}
	}

	if status, raw, err := s.nodeGET(ctx, base+"/api/databases/health", token); err == nil && status == 200 {
		var rows []struct {
			Name      string `json:"name"`
			Overall   string `json:"overall"`
			Message   string `json:"message"`
			Databases []struct {
				Name    string `json:"name"`
				Overall string `json:"overall"`
				Message string `json:"message"`
			} `json:"databases"`
		}
		if json.Unmarshal(raw, &rows) == nil {
			for _, row := range rows {
				if len(row.Databases) == 0 {
					if row.Overall == "healthy" {
						continue
					}
					name := row.Name
					if name == "" {
						name = "Postgres"
					}
					fmt.Fprintf(&issues, "%s База данных: %s — %s", remoteStatusEmoji(row.Overall), html.EscapeString(name), remoteOverallLabel(row.Overall))
					if msg := strings.TrimSpace(row.Message); msg != "" {
						fmt.Fprintf(&issues, ": %s", html.EscapeString(msg))
					}
					issues.WriteByte('\n')
					continue
				}
				for _, dbRow := range row.Databases {
					if dbRow.Overall == "healthy" {
						continue
					}
					name := row.Name
					if name == "" {
						name = "Postgres"
					}
					name = name + " / " + dbRow.Name
					fmt.Fprintf(&issues, "%s База данных: %s — %s", remoteStatusEmoji(dbRow.Overall), html.EscapeString(name), remoteOverallLabel(dbRow.Overall))
					if msg := strings.TrimSpace(dbRow.Message); msg != "" {
						fmt.Fprintf(&issues, ": %s", html.EscapeString(msg))
					}
					issues.WriteByte('\n')
				}
			}
		}
	}

	header := fmt.Sprintf("<b>Barn — %s</b>\n", html.EscapeString(panelName))
	if issues.Len() > 0 {
		return header + strings.TrimSpace(issues.String()), nil
	}
	return header + "✅ Всё ОК", nil
}

func (s *Service) nodeGET(ctx context.Context, url, token string) (int, []byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-Barn-Token", token)
	client := &http.Client{Timeout: 20 * time.Second, CheckRedirect: limitRedirects}
	res, err := client.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer res.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	return res.StatusCode, raw, nil
}

func remoteStatusEmoji(overall string) string {
	switch overall {
	case "degraded":
		return "⚠️"
	case "unhealthy":
		return "❌"
	default:
		return "❔"
	}
}

func remoteOverallLabel(overall string) string {
	switch overall {
	case "degraded":
		return "есть проблемы"
	case "unhealthy":
		return "недоступен"
	default:
		return "статус неизвестен"
	}
}
