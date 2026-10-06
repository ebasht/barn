package servers

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
)

const proxyMaxBody = 512 << 20 // 512 MiB (DB restore uploads)

// AllowedProxyPath reports whether Master may forward path (relative to /api/) to a node.
func AllowedProxyPath(apiPath string) bool {
	p := strings.TrimPrefix(strings.TrimSpace(apiPath), "/")
	if p == "" {
		return false
	}
	blocked := []string{
		"servers/node",
		"servers/ingest",
		"servers/agent",
		"servers/nodes",
		"servers/installations",
		"auth/",
	}
	for _, b := range blocked {
		if p == strings.TrimSuffix(b, "/") || strings.HasPrefix(p, b) {
			return false
		}
	}
	allowed := []string{
		"sites",
		"databases",
		"backups",
		"billing",
		"notifications",
		"system",
		"deployments",
		"mcp",
		"servers/settings",
		"servers/pairing-code",
		"servers/master",
		"servers/overview",
		"servers/events",
		"servers/incidents",
	}
	for _, a := range allowed {
		if p == a || strings.HasPrefix(p, a+"/") {
			return true
		}
	}
	return false
}

// ProxyNodeAPI forwards an authenticated Master request to a paired Barn panel.
func (s *Service) ProxyNodeAPI(w http.ResponseWriter, r *http.Request, nodeID uuid.UUID, apiPath string) {
	settings, err := s.ensureSettings(r.Context())
	if err != nil {
		writeProxyError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if settings.Mode != ModeMaster {
		writeProxyError(w, http.StatusForbidden, ErrForbidden.Error())
		return
	}
	if !AllowedProxyPath(apiPath) {
		writeProxyError(w, http.StatusBadRequest, "path not allowed for remote proxy")
		return
	}
	node, err := s.q.GetServersNode(r.Context(), nodeID)
	if err != nil {
		writeProxyError(w, http.StatusNotFound, "node not found")
		return
	}
	if !IsBarnPanel(node.ConnectionType) {
		writeProxyError(w, http.StatusBadRequest, "node is not a Barn panel")
		return
	}
	base := strings.TrimRight(strings.TrimSpace(node.BaseUrl), "/")
	if base == "" {
		writeProxyError(w, http.StatusBadRequest, "node has no base URL")
		return
	}
	cred, err := s.q.GetOutboundCredential(r.Context(), node.ID)
	if err != nil || len(cred.EncryptedToken) == 0 {
		writeProxyError(w, http.StatusBadGateway, "missing remote credentials")
		return
	}
	token, err := s.cipher.Decrypt(cred.EncryptedToken)
	if err != nil || strings.TrimSpace(token) == "" {
		writeProxyError(w, http.StatusBadGateway, "cannot decrypt remote credentials")
		return
	}

	target, err := url.Parse(base + "/api/" + strings.TrimPrefix(apiPath, "/"))
	if err != nil {
		writeProxyError(w, http.StatusBadRequest, "invalid remote URL")
		return
	}
	q := r.URL.Query()
	q.Del("token")
	target.RawQuery = q.Encode()

	var body io.Reader = http.NoBody
	if r.Body != nil && (r.ContentLength > 0 || r.ContentLength == -1) {
		body = io.LimitReader(r.Body, proxyMaxBody+1)
	}

	ctx := r.Context()
	var cancel context.CancelFunc
	if !isStreamingProxy(r, apiPath) {
		ctx, cancel = context.WithTimeout(ctx, 120*time.Second)
		defer cancel()
	}

	req, err := http.NewRequestWithContext(ctx, r.Method, target.String(), body)
	if err != nil {
		writeProxyError(w, http.StatusBadGateway, "proxy request failed")
		return
	}
	if r.ContentLength > 0 {
		req.ContentLength = r.ContentLength
	}
	if ct := r.Header.Get("Content-Type"); ct != "" {
		req.Header.Set("Content-Type", ct)
	}
	if accept := r.Header.Get("Accept"); accept != "" {
		req.Header.Set("Accept", accept)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-Barn-Token", token)

	client := &http.Client{
		Timeout: 0,
		CheckRedirect: func(_ *http.Request, via []*http.Request) error {
			if len(via) >= 5 {
				return fmt.Errorf("too many redirects")
			}
			return nil
		},
	}
	res, err := client.Do(req)
	if err != nil {
		writeProxyError(w, http.StatusBadGateway, "remote request failed: "+err.Error())
		return
	}
	defer res.Body.Close()

	for k, vals := range res.Header {
		switch strings.ToLower(k) {
		case "connection", "keep-alive", "transfer-encoding", "proxy-authenticate",
			"proxy-authorization", "te", "trailers", "upgrade":
			continue
		}
		for _, v := range vals {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(res.StatusCode)
	if flusher, ok := w.(http.Flusher); ok {
		buf := make([]byte, 32*1024)
		for {
			n, readErr := res.Body.Read(buf)
			if n > 0 {
				if _, writeErr := w.Write(buf[:n]); writeErr != nil {
					return
				}
				flusher.Flush()
			}
			if readErr != nil {
				return
			}
		}
	}
	_, _ = io.Copy(w, io.LimitReader(res.Body, proxyMaxBody))
}

func isStreamingProxy(r *http.Request, apiPath string) bool {
	if strings.Contains(apiPath, "/stream") {
		return true
	}
	return strings.Contains(strings.ToLower(r.Header.Get("Accept")), "text/event-stream")
}

func writeProxyError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write([]byte(`{"error":` + jsonQuote(msg) + `}`))
}

func jsonQuote(s string) string {
	var b strings.Builder
	b.WriteByte('"')
	for _, r := range s {
		switch r {
		case '\\', '"':
			b.WriteByte('\\')
			b.WriteRune(r)
		case '\n':
			b.WriteString(`\n`)
		case '\r':
			b.WriteString(`\r`)
		case '\t':
			b.WriteString(`\t`)
		default:
			if r < 0x20 {
				b.WriteString(fmt.Sprintf(`\u%04x`, r))
				continue
			}
			b.WriteRune(r)
		}
	}
	b.WriteByte('"')
	return b.String()
}
