package barnmcp

import (
	"context"
	"crypto/sha256"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// Memory DBTX exercises service queries and API-visible behavior without a live DB.
type settingsDB struct {
	value     Settings
	hash      []byte
	exists    bool
	failWrite bool
}
type settingsRow struct{ store *settingsDB }

func (r settingsRow) Scan(dest ...any) error {
	s := r.store
	if !s.exists {
		return pgx.ErrNoRows
	}
	*dest[0].(*string) = s.value.InstanceID
	*dest[1].(*string) = s.value.InstanceName
	*dest[2].(*[]byte) = append([]byte(nil), s.hash...)
	*dest[3].(*bool) = s.value.AllowWrites
	*dest[4].(**time.Time) = s.value.KeyCreatedAt
	return nil
}
func (s *settingsDB) QueryRow(context.Context, string, ...any) pgx.Row { return settingsRow{s} }
func (s *settingsDB) Query(context.Context, string, ...any) (pgx.Rows, error) {
	return nil, errors.New("unexpected query")
}
func (s *settingsDB) Exec(_ context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	if s.failWrite {
		return pgconn.CommandTag{}, errors.New("database write failed")
	}
	switch {
	case strings.HasPrefix(sql, "INSERT"):
		if !s.exists {
			s.exists = true
			s.value.InstanceID = args[0].(string)
			s.value.InstanceName = args[1].(string)
			s.hash = append([]byte(nil), args[2].([]byte)...)
			s.value.AllowWrites = args[3].(bool)
			s.value.KeyCreatedAt = args[4].(*time.Time)
		}
	case strings.Contains(sql, "instance_name=$1"):
		s.value.InstanceName = args[0].(string)
		s.value.AllowWrites = args[1].(bool)
	case strings.Contains(sql, "token_hash=$1"):
		s.hash = append([]byte(nil), args[0].([]byte)...)
		now := time.Now()
		s.value.KeyCreatedAt = &now
	case strings.Contains(sql, "token_hash=NULL"):
		s.hash = nil
		s.value.KeyCreatedAt = nil
	default:
		return pgconn.CommandTag{}, errors.New("unexpected statement")
	}
	return pgconn.NewCommandTag("UPDATE 1"), nil
}

func TestKeyLifecycle(t *testing.T) {
	ctx := context.Background()
	store := &settingsDB{}
	service := NewSettingsService(store, "", false, InstanceInfo{Name: "Main"})
	initial, err := service.Get(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if initial.InstanceID == "" || initial.KeyConfigured {
		t.Fatalf("%+v", initial)
	}
	handler := ManagedHandler(service.Access, nil, nil, nil, nil)
	check := func(token string, wantStatus int, writes bool) {
		t.Helper()
		req := httptest.NewRequest("POST", "/mcp", strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"tools/list"}`))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json, text/event-stream")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		if w.Code != wantStatus {
			t.Fatalf("status %d want %d: %s", w.Code, wantStatus, w.Body.String())
		}
		if wantStatus == 200 && strings.Contains(w.Body.String(), `"name":"deploy_site"`) != writes {
			t.Fatalf("wrong permissions: %s", w.Body.String())
		}
	}
	check("", 404, false)
	first, err := service.Generate(ctx)
	if err != nil {
		t.Fatal(err)
	}
	hash := sha256.Sum256([]byte(first))
	if string(store.hash) != string(hash[:]) {
		t.Fatal("only SHA-256 hash must be persisted")
	}
	check(first, 200, false)
	if _, err := service.Update(ctx, "Production", true); err != nil {
		t.Fatal(err)
	}
	check(first, 200, true)
	second, err := service.Generate(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if first == second {
		t.Fatal("rotation reused key")
	}
	check(first, 401, false)
	check(second, 200, true)
	restarted := NewSettingsService(store, "legacy-environment-token", false, InstanceInfo{ID: "other", Name: "Other"})
	if err := restarted.Ensure(ctx); err != nil {
		t.Fatal(err)
	}
	persisted, err := restarted.Get(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if persisted.InstanceID != initial.InstanceID || persisted.InstanceName != "Production" || !persisted.AllowWrites {
		t.Fatalf("settings overwritten on restart: %+v", persisted)
	}
	if err := service.Revoke(ctx); err != nil {
		t.Fatal(err)
	}
	check(second, 404, false)
	if err := restarted.Ensure(ctx); err != nil {
		t.Fatal(err)
	}
	after, err := restarted.Get(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if after.KeyConfigured {
		t.Fatal("environment token reenabled revoked access")
	}
}

func TestSettingsValidationAndFailedGeneration(t *testing.T) {
	ctx := context.Background()
	store := &settingsDB{}
	service := NewSettingsService(store, "", false, InstanceInfo{ID: "main", Name: "Main"})
	if _, err := service.Update(ctx, " ", true); !errors.Is(err, ErrInvalidSettings) {
		t.Fatal("empty name accepted")
	}
	if _, err := service.Get(ctx); err != nil {
		t.Fatal(err)
	}
	store.failWrite = true
	token, err := service.Generate(ctx)
	if err == nil || token != "" {
		t.Fatal("failed write returned usable-looking key")
	}
}
