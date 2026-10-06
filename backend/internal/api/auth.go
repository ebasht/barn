package api

import (
	"net/http"
	"strings"

	"crypto/subtle"

	"github.com/ebash/barn/backend/internal/servers"
)

const (
	headerAuthorization = "Authorization"
	headerAPIToken      = "X-API-Token"
	queryToken          = "token"
)

// BearerTokenAuth protects routes with a static API token from config.
// Accepts: Authorization: Bearer <token>, X-API-Token: <token>, or ?token= (for SSE).
func BearerTokenAuth(expected string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !tokenValid(r, expected) {
				writeJSON(w, http.StatusUnauthorized, errorBody{Error: "unauthorized"})
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// PanelAuth accepts the panel API token or a Master→node panel-admin credential.
func PanelAuth(expected string, serversSvc *servers.Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if tokenValid(r, expected) {
				next.ServeHTTP(w, r)
				return
			}
			if serversSvc != nil && serversSvc.PanelAdminAuthorized(r) {
				next.ServeHTTP(w, r)
				return
			}
			writeJSON(w, http.StatusUnauthorized, errorBody{Error: "unauthorized"})
		})
	}
}

func tokenValid(r *http.Request, expected string) bool {
	if expected == "" {
		return false
	}

	if token, ok := bearerToken(r.Header.Get(headerAuthorization)); ok {
		return secureEqual(token, expected)
	}

	if t := r.Header.Get(headerAPIToken); t != "" {
		return secureEqual(t, expected)
	}

	if t := r.URL.Query().Get(queryToken); t != "" {
		return secureEqual(t, expected)
	}

	return false
}

func bearerToken(header string) (string, bool) {
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return "", false
	}
	token := strings.TrimSpace(header[len(prefix):])
	return token, token != ""
}

func secureEqual(got, want string) bool {
	return subtle.ConstantTimeCompare([]byte(got), []byte(want)) == 1
}
