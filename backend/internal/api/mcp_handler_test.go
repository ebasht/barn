package api

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestMCPKeyManagementRequiresPanelToken(t *testing.T) {
	panelToken := strings.Repeat("p", 32)
	router := NewRouter(Handlers{MCP: NewMCPHandler(nil)}, panelToken, nil)
	for _, route := range []struct{ method, path string }{{"GET", "/api/mcp/settings"}, {"PUT", "/api/mcp/settings"}, {"POST", "/api/mcp/key"}, {"DELETE", "/api/mcp/key"}} {
		req := httptest.NewRequest(route.method, route.path, nil)
		req.Header.Set("Authorization", "Bearer barn_mcp_wrong")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != 401 {
			t.Errorf("%s %s: got %d", route.method, route.path, w.Code)
		}
	}
}
