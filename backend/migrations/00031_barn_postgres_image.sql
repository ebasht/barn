-- +goose Up
-- Canonical image name is barn-postgres (built FROM pgvector in release).
ALTER TABLE pdb_instances ALTER COLUMN image SET DEFAULT 'barn-postgres:latest';

UPDATE pdb_instances
SET image = 'barn-postgres:latest', updated_at = now()
WHERE lower(image) IN (
    'postgres:16-alpine',
    'postgres:16',
    'postgres:16-bookworm',
    'postgres:16-trixie',
    'pgvector/pgvector:pg16'
)
   OR (
        lower(image) LIKE 'postgres:16.%'
        AND lower(image) NOT LIKE '%vector%'
   );

-- +goose Down
ALTER TABLE pdb_instances ALTER COLUMN image SET DEFAULT 'pgvector/pgvector:pg16';
