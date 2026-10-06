package servers

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/ebash/barn/backend/internal/db"
)

type ctxKey int

const (
	ctxNodeID ctxKey = iota + 1
	ctxScopes
)

func NodeIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(ctxNodeID).(uuid.UUID)
	return id, ok
}

func ScopesFromContext(ctx context.Context) []string {
	s, _ := ctx.Value(ctxScopes).([]string)
	return s
}

func (s *Service) ServersTokenAuth(requiredScope string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			raw := extractBearer(r)
			if raw == "" {
				writeServersUnauthorized(w)
				return
			}
			cred, err := s.q.GetInboundCredentialByHash(r.Context(), HashToken(raw))
			if err != nil {
				writeServersUnauthorized(w)
				return
			}
			if !tokenHashEqual(cred.TokenHash, HashToken(raw)) {
				writeServersUnauthorized(w)
				return
			}
			if requiredScope != "" && !HasScope(cred.Scopes, requiredScope) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte(`{"error":"forbidden"}`))
				return
			}
			node, err := s.q.GetServersNode(r.Context(), cred.NodeID)
			if err != nil || node.DeletedAt.Valid {
				writeServersUnauthorized(w)
				return
			}
			ctx := context.WithValue(r.Context(), ctxNodeID, cred.NodeID)
			ctx = context.WithValue(ctx, ctxScopes, append([]string(nil), cred.Scopes...))
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func writeServersUnauthorized(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_, _ = w.Write([]byte(`{"error":"unauthorized"}`))
}

func extractBearer(r *http.Request) string {
	h := r.Header.Get("Authorization")
	const prefix = "Bearer "
	if strings.HasPrefix(h, prefix) {
		return strings.TrimSpace(h[len(prefix):])
	}
	if t := strings.TrimSpace(r.Header.Get("X-Barn-Token")); t != "" {
		return t
	}
	// Legacy header name (pre /api/servers rename)
	return strings.TrimSpace(r.Header.Get("X-Fleet-Token"))
}

// storeNodeInboundToken saves hash of the token Master will use to call this node.
func (s *Service) storeNodeInboundToken(ctx context.Context, nodeToken string) error {
	settings, err := s.ensureSettings(ctx)
	if err != nil {
		return err
	}
	local, err := s.q.GetServersNodeByUID(ctx, settings.NodeUid)
	if err != nil {
		caps, _ := json.Marshal(BarnCapabilities())
		now := time.Now().UTC()
		local, err = s.q.CreateServersNode(ctx, db.CreateServersNodeParams{
			NodeUid:        settings.NodeUid,
			Name:           settings.NodeName,
			Role:           RoleNode,
			ConnectionType: ConnLocal,
			BaseUrl:        settings.PublicUrl,
			Status:         StatusOnline,
			Capabilities:   caps,
			Version:        s.appVersion,
			AgentVersion:   "",
			LastSeenAt:     pgTimestamptz(now),
			PairedAt:       pgTimestamptz(now),
			Metadata:       []byte("{}"),
		})
		if err != nil {
			return mapErr(err)
		}
	}
	_, err = s.q.CreateServersCredential(ctx, db.CreateServersCredentialParams{
		NodeID:         local.ID,
		Direction:      "inbound",
		Purpose:        "master_to_node",
		Scopes:         MasterToNodeScopes(),
		TokenHash:      HashToken(nodeToken),
		EncryptedToken: nil,
	})
	return mapErr(err)
}

// PanelAdminAuthorized reports whether the request carries a Master→node token
// allowed to use the full panel API (sites, databases, backups, …).
func (s *Service) PanelAdminAuthorized(r *http.Request) bool {
	raw := extractPanelToken(r)
	if raw == "" {
		return false
	}
	cred, err := s.q.GetInboundCredentialByHash(r.Context(), HashToken(raw))
	if err != nil {
		return false
	}
	if !tokenHashEqual(cred.TokenHash, HashToken(raw)) {
		return false
	}
	if cred.Purpose != "master_to_node" {
		return false
	}
	// Require panel:admin when present; accept legacy master_to_node read tokens
	// until EnsurePanelAdminScopes rewrites them on the managed node.
	if !HasScope(cred.Scopes, ScopePanelAdmin) &&
		!HasScope(cred.Scopes, "*") &&
		!HasScope(cred.Scopes, ScopeStatusRead) {
		return false
	}
	node, err := s.q.GetServersNode(r.Context(), cred.NodeID)
	if err != nil || node.DeletedAt.Valid {
		return false
	}
	return true
}

func extractPanelToken(r *http.Request) string {
	if raw := extractBearer(r); raw != "" {
		return raw
	}
	if t := strings.TrimSpace(r.Header.Get(headerAPITokenLegacy)); t != "" {
		return t
	}
	return strings.TrimSpace(r.URL.Query().Get("token"))
}

const headerAPITokenLegacy = "X-API-Token"
