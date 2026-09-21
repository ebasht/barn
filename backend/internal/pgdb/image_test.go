package pgdb

import "testing"

func TestNormalizeManagedPostgresImage(t *testing.T) {
	t.Parallel()
	cases := map[string]string{
		"":                       DefaultManagedPostgresImage,
		"postgres:16-alpine":     DefaultManagedPostgresImage,
		"postgres:16":            DefaultManagedPostgresImage,
		"Postgres:16-Alpine":     DefaultManagedPostgresImage,
		"postgres:16.6-alpine":   DefaultManagedPostgresImage,
		"pgvector/pgvector:pg16": DefaultManagedPostgresImage,
		"barn-postgres:latest":   DefaultManagedPostgresImage,
		"myregistry/pg:custom":   "myregistry/pg:custom",
	}
	for in, want := range cases {
		if got := NormalizeManagedPostgresImage(in); got != want {
			t.Fatalf("NormalizeManagedPostgresImage(%q)=%q want %q", in, got, want)
		}
	}
}
