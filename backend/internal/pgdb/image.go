package pgdb

import (
	"context"
	"fmt"
	"strings"
)

// DefaultManagedPostgresImage is the stock managed Postgres image shipped in
// Barn releases (docker/postgres/Dockerfile — Postgres 16 + pgvector).
const DefaultManagedPostgresImage = "barn-postgres:latest"

// upstreamPgvectorImage is used only when barn-postgres is not present locally
// (e.g. Deploy before a release image was loaded).
const upstreamPgvectorImage = "pgvector/pgvector:pg16"

// NormalizeManagedPostgresImage upgrades known stock Postgres 16 images that
// lack pgvector to DefaultManagedPostgresImage. Custom / offline tags are left
// unchanged.
func NormalizeManagedPostgresImage(image string) string {
	image = strings.TrimSpace(image)
	if image == "" {
		return DefaultManagedPostgresImage
	}
	lower := strings.ToLower(image)
	switch lower {
	case "postgres:16-alpine", "postgres:16", "postgres:16-bookworm", "postgres:16-trixie",
		"pgvector/pgvector:pg16", "dock-pilot-postgres:latest":
		return DefaultManagedPostgresImage
	}
	if strings.HasPrefix(lower, "postgres:16.") && !strings.Contains(lower, "vector") {
		return DefaultManagedPostgresImage
	}
	return image
}

// ensureManagedImage makes image available locally. barn-postgres is release-local;
// if missing, materialize it from upstream pgvector.
func (s *Service) ensureManagedImage(ctx context.Context, image string) error {
	image = strings.TrimSpace(image)
	if image == "" {
		image = DefaultManagedPostgresImage
	}
	if s.docker.ImageExists(ctx, image) {
		return nil
	}
	if err := s.docker.Pull(ctx, image); err == nil {
		return nil
	} else if image != DefaultManagedPostgresImage && image != "dock-pilot-postgres:latest" {
		return err
	} else {
		s.logger.InfoContext(ctx, "local postgres image missing; pulling upstream pgvector",
			"wanted", image, "upstream", upstreamPgvectorImage, "pull_error", err)
	}
	if err := s.docker.Pull(ctx, upstreamPgvectorImage); err != nil {
		return fmt.Errorf("pull %s: %w", upstreamPgvectorImage, err)
	}
	if _, err := s.docker.ResolveLocalImage(ctx, image, upstreamPgvectorImage); err != nil {
		return fmt.Errorf("tag %s as %s: %w", upstreamPgvectorImage, image, err)
	}
	return nil
}
