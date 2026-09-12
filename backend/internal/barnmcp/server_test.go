package barnmcp

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ebash/barn/backend/internal/sites"
)

func TestEndpoint(t *testing.T) {
	token := strings.Repeat("x", 32)
	for _, tc := range []struct {
		name, token, auth, origin string
		writes                    bool
		status                    int
	}{
		{"disabled", "", "", "", false, 404},
		{"unauthorized", token, "", "", false, 401},
		{"wrong token", token, "Bearer wrong", "", false, 401},
		{"query token rejected", token, "", "", false, 401},
		{"origin rejected", token, "Bearer " + token, "https://evil.example", false, 403},
		{"read tools", token, "Bearer " + token, "", false, 200},
		{"write tools", token, "Bearer " + token, "https://barn.example", true, 200},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h := Handler(tc.token, tc.writes, []string{"https://barn.example"}, nil, nil, nil, InstanceInfo{ID: "main", Name: "Main panel"})
			req := httptest.NewRequest("POST", "/mcp?token="+token, strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"tools/list"}`))
			req.Header.Set("Authorization", tc.auth)
			req.Header.Set("Origin", tc.origin)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Accept", "application/json, text/event-stream")
			w := httptest.NewRecorder()
			h.ServeHTTP(w, req)
			if w.Code != tc.status {
				t.Fatalf("status %d: %s", w.Code, w.Body.String())
			}
			if tc.status == 200 {
				body := w.Body.String()
				if !strings.Contains(body, "list_sites") {
					t.Fatalf("missing tools: %s", body)
				}
				if strings.Contains(body, "deploy_site") != tc.writes {
					t.Fatal("write tool visibility")
				}
			}
		})
	}
}

func TestInitialize(t *testing.T) {
	token := strings.Repeat("x", 32)
	h := Handler(token, false, nil, nil, nil, nil, InstanceInfo{ID: "main", Name: "Main panel"})
	req := httptest.NewRequest("POST", "/mcp", strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-11-25","capabilities":{},"clientInfo":{"name":"test","version":"1"}}}`))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("%d %s", w.Code, w.Body.String())
	}
	// Streamable HTTP may return JSON inside an SSE data event.
	var response map[string]any
	body := w.Body.String()
	if strings.Contains(body, "data: ") {
		body = strings.Split(strings.SplitN(body, "data: ", 2)[1], "\n")[0]
	}
	if err := json.Unmarshal([]byte(body), &response); err != nil {
		t.Fatal(err)
	}
	if response["result"] == nil {
		t.Fatalf("%s", w.Body.String())
	}
}

func TestMatchingSites(t *testing.T) {
	items := []sites.SiteListItem{{Name: "Писарь", Slug: "pisar"}, {Name: "Писарь", Slug: "pisar-copy"}, {Name: "Писарь тест", Slug: "pisar-test"}}
	for _, tc := range []struct {
		query string
		want  int
	}{{" писарь ", 2}, {"PISAR", 1}, {"пис", 0}, {"unknown", 0}, {"", 3}} {
		if got := len(matchingSites(items, tc.query)); got != tc.want {
			t.Errorf("query %q: got %d want %d", tc.query, got, tc.want)
		}
	}
}

func TestValidateTarget(t *testing.T) {
	panel := InstanceInfo{ID: "main"}
	for _, tc := range []struct {
		instance, name string
		valid          bool
	}{{"main", "Писарь", true}, {"main", " PISAR ", true}, {"side", "Писарь", false}, {"", "Писарь", false}, {"main", "", false}, {"main", "Писарь тест", false}} {
		err := validateTarget(panel, Input{InstanceID: tc.instance, SiteName: tc.name}, "Писарь", "pisar")
		if (err == nil) != tc.valid {
			t.Errorf("%+v: %v", tc, err)
		}
	}
}

func TestInstanceInfoAndWrongPanelWrite(t *testing.T) {
	token := strings.Repeat("x", 32)
	h := Handler(token, true, nil, nil, nil, nil, InstanceInfo{ID: "main", Name: "Main panel"})
	for _, tc := range []struct{ body, want string }{
		{`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"get_instance_info","arguments":{}}}`, "Main panel"},
		{`{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"deploy_site","arguments":{"instance_id":"side","site_name":"Писарь"}}}`, "instance_id must match"},
	} {
		req := httptest.NewRequest("POST", "/mcp", strings.NewReader(tc.body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json, text/event-stream")
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		if w.Code != 200 || !strings.Contains(w.Body.String(), tc.want) {
			t.Fatalf("%d: %s", w.Code, w.Body.String())
		}
	}
}
