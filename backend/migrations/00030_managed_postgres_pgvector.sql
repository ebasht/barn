-- +goose Up
-- Managed Postgres defaults to an image with pgvector preinstalled.
-- Updating the column alone does not recreate the container; redeploy /
-- barn-upgrade keeps the existing data volume and swaps only the image.
ALTER TABLE pdb_instances ALTER COLUMN image SET DEFAULT 'pgvector/pgvector:pg16';

UPDATE pdb_instances
SET image = 'pgvector/pgvector:pg16', updated_at = now()
WHERE lower(image) IN (
    'postgres:16-alpine',
    'postgres:16',
    'postgres:16-bookworm',
    'postgres:16-trixie'
)
   OR (
        lower(image) LIKE 'postgres:16.%'
        AND lower(image) NOT LIKE '%vector%'
   );

-- +goose Down
ALTER TABLE pdb_instances ALTER COLUMN image SET DEFAULT 'postgres:16-alpine';
