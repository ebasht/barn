-- +goose Up
ALTER TABLE notification_settings ADD COLUMN daily_digest_mode TEXT NOT NULL DEFAULT 'detailed' CHECK (daily_digest_mode IN ('detailed', 'compact'));

-- +goose Down
ALTER TABLE notification_settings DROP COLUMN daily_digest_mode;
