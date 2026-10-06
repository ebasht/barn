package pgdb

import (
	"net/url"
	"os"
	"testing"
)

func TestPanelPostgresHostPort(t *testing.T) {
	t.Setenv("POSTGRES_HOST_PORT", "")
	t.Setenv("DATABASE_URL", "")
	if got := panelPostgresHostPort(); got != 5433 {
		t.Fatalf("default: got %d", got)
	}

	t.Setenv("POSTGRES_HOST_PORT", "5544")
	if got := panelPostgresHostPort(); got != 5544 {
		t.Fatalf("env port: got %d", got)
	}

	t.Setenv("POSTGRES_HOST_PORT", "")
	t.Setenv("DATABASE_URL", "postgres://barn:x@127.0.0.1:5999/barn?sslmode=disable")
	if got := panelPostgresHostPort(); got != 5999 {
		t.Fatalf("database url: got %d", got)
	}

	// Invalid POSTGRES_HOST_PORT falls through to DATABASE_URL.
	t.Setenv("POSTGRES_HOST_PORT", "nope")
	u := &url.URL{Scheme: "postgres", Host: "127.0.0.1:6001", Path: "/barn"}
	_ = os.Setenv("DATABASE_URL", u.String())
	if got := panelPostgresHostPort(); got != 6001 {
		t.Fatalf("fallback url: got %d", got)
	}
}
