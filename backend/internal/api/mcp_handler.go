package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/ebash/barn/backend/internal/barnmcp"
)

type MCPHandler struct{ service *barnmcp.SettingsService }

func NewMCPHandler(s *barnmcp.SettingsService) *MCPHandler { return &MCPHandler{s} }
func (h *MCPHandler) Get(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	v, err := h.service.Get(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, v)
}
func (h *MCPHandler) Update(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	var body struct {
		Name        string `json:"instance_name"`
		AllowWrites bool   `json:"allow_writes"`
	}
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&body); err != nil {
		writeJSON(w, 400, errorBody{Error: "invalid MCP settings"})
		return
	}
	v, err := h.service.Update(r.Context(), body.Name, body.AllowWrites)
	if errors.Is(err, barnmcp.ErrInvalidSettings) {
		writeJSON(w, 400, errorBody{Error: err.Error()})
		return
	}
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, 200, v)
}
func (h *MCPHandler) Generate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	token, err := h.service.Generate(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, 200, struct {
		Token string `json:"token"`
	}{token})
}
func (h *MCPHandler) Revoke(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	if err := h.service.Revoke(r.Context()); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
