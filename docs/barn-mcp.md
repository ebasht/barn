# Barn MCP

Barn exposes Streamable HTTP at `/mcp`, in the API process. No local adapter or AI key is required.

## Enable

Apply backend migration `00028_mcp_settings.sql` with the normal Barn migration/upgrade process, then update the API and frontend.

Open **Settings → MCP access** (`/servers/settings`):

1. Set a recognizable panel name and save.
2. Enable deployments/restarts if needed and save. Read access is the default.
3. Click **Generate key** and copy it immediately. It is shown once and is not stored in browser storage.
4. Copy the MCP URL/Codex configuration into your AI client.

The key works immediately without restarting Barn. Replacing a key invalidates the previous key on subsequent requests. Revoking it disables MCP (404). Requests already in flight may finish. MCP settings persist in PostgreSQL; only the SHA-256 hash of the key is stored. The key grants access to all sites on this instance; per-user/site scopes are not implemented yet. Settings/key endpoints use the existing panel API authentication and do not accept MCP keys. Only the Authorization Bearer header is accepted at `/mcp`.

Legacy `BARN_MCP_TOKEN`, `BARN_MCP_ALLOW_WRITES`, `BARN_MCP_INSTANCE_ID` and `BARN_MCP_INSTANCE_NAME` seed the settings only on first initialization. Subsequent restarts/environment changes do not overwrite panel settings or restore a revoked key. Without an explicit or legacy instance ID, a stable UUID is generated and persisted. Reconnect your AI client after changing permissions to refresh its tool list; revoked write permissions also apply to calls made using a cached tool list.

The panel nginx templates proxy `/mcp` to the API. Existing installations need the same location added to their active panel server block and nginx configuration tested/reloaded:

```nginx
location = /mcp {
    proxy_pass http://127.0.0.1:8080; # replace with your API port
    proxy_http_version 1.1;
    proxy_set_header Host $host;
    proxy_buffering off;
    proxy_cache off;
    proxy_read_timeout 300s;
}
```

Use HTTPS for remote connections. An Origin header, when supplied, must match `CORS_ALLOWED_ORIGINS`.

## Connect Codex

Set `BARN_MCP_TOKEN` in the environment of the Codex process, then add to your Codex configuration:

```toml
[mcp_servers.barn]
url = "https://barn.example.com/mcp"
bearer_token_env_var = "BARN_MCP_TOKEN"
```

Restart Codex after changing its environment/configuration. See [official Codex MCP documentation](https://developers.openai.com/codex/mcp).

## Tools

Read: `get_instance_info`, `list_sites`, `get_health`, `get_system_status`, `get_logs`, `list_deployments`, `get_deployment`, `get_deployment_logs`.

Optional writes: `deploy_site`, `restart_site`. Supply `site_id`, `site_name` and `instance_id` from `list_sites` for writes. The server rejects a wrong instance ID or site name. Deployment status/log tools accept `deployment_id`. Deployment logs accept `after_id` and `limit` (default 100, max 500), returning `next_after_id` and `has_more`. Container logs return history without following (default 100, max 500).

Deployment uses the site's configured branch and returns immediately with a job ID. Check status and logs, then health. Deployment requests are not idempotent: do not blindly retry after a timeout; check history first. MCP tool annotations are client hints, not approval enforcement. Configure approval in your AI client as needed.

Application/build logs can contain sensitive data and untrusted text. Env variable values, secrets management, arbitrary shell execution, deletion and restore are not exposed by these tools.

## Multiple panels and site names

Connect each panel as a separate MCP server in Codex. Give each panel a recognizable name in its MCP settings. Each panel has a stable database-backed instance ID, generated automatically unless seeded from legacy environment settings. Existing explicit IDs must remain unique across connected panels.

`get_instance_info` returns the panel identity. `list_sites` returns `{instance, sites}` and accepts an optional `query` for an exact name or slug match, ignoring case/whitespace. Duplicate names are returned without choosing a winner.

For “deploy Писарь”, MCP instructions tell the AI to search every connected Barn panel, select a unique matching site, and deploy through the same panel with its instance ID, site UUID and name. If multiple sites match, or a panel is unavailable, it should ask for the target. If the user specifies a panel, only that panel needs searching. Status and health checks stay on the selected panel.

Each MCP server sees only its own panel. Cross-panel discovery and ambiguity handling are performed by the AI client; the server cannot enforce that every other panel was searched. It does validate the supplied panel ID and site name before a write. Actual client routing should be verified with the connected panels before relying on it.
