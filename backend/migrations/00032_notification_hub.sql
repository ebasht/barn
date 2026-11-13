-- +goose Up
ALTER TABLE servers_settings
    ADD COLUMN IF NOT EXISTS notification_hub_node_id UUID
        REFERENCES servers_nodes(id) ON DELETE SET NULL;

-- +goose Down
ALTER TABLE servers_settings
    DROP COLUMN IF EXISTS notification_hub_node_id;
