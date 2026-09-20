package pgdb

import "strings"

// DefaultManagedPostgresImage is the stock managed Postgres image.
// Includes pgvector so CREATE EXTENSION vector works out of the box.
const DefaultManagedPostgresImage = "pgvector/pgvector:pg16"

// NormalizeManagedPostgresImage upgrades known stock Postgres 16 images that
// lack pgvector to DefaultManagedPostgresImage. Custom / offline tags
// (barn-postgres:latest, etc.) are left unchanged.
func NormalizeManagedPostgresImage(image string) string {
	image = strings.TrimSpace(image)
	if image == "" {
		return DefaultManagedPostgresImage
	}
	lower := strings.ToLower(image)
	switch lower {
	case "postgres:16-alpine", "postgres:16", "postgres:16-bookworm", "postgres:16-trixie":
		return DefaultManagedPostgresImage
	}
	if strings.HasPrefix(lower, "postgres:16.") && !strings.Contains(lower, "vector") {
		return DefaultManagedPostgresImage
	}
	return image
}
