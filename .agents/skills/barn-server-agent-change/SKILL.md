---
name: barn-server-agent-change
description: >
  Change Barn master, managed-node, pairing, scoped tokens, monitoring agent,
  heartbeat/events, SSH installation, systemd, builds, and DockPilot compatibility.
---

# Barn Server and Agent Change

Use this workflow for server management, pairing, node polling, agent registration/update, heartbeat/events, SSH install jobs, systemd units, or agent artifacts.

## Establish the actual mode and lifecycle

Read `AGENTS.md`, `backend/AGENTS.md`, `scripts/AGENTS.md` when packaging/install is affected, `docs/servers.md`, `docs/barn-agent.md`, and the relevant code under `backend/internal/servers`, `backend/cmd/agent`, `backend/internal/agent`, and `backend/internal/metrics`. Inspect router middleware, migrations/queries, frontend server pages/API/types, Dockerfile, Makefile, service units, and release scripts for cross-cutting changes.

Barn modes are `standalone`, `master`, and `managed_node`. Preserve their boundaries:

- standalone remains the default and retains local sites/settings;
- master manages its local server plus remote Barn panels and monitoring agents;
- a managed node is paired to one master and can forward centralized notification events;
- nesting/pairing conflicts and master-only operations are enforced by the service, not just hidden in UI.

## Protect trust boundaries

- Pairing codes and registration tokens are short-lived, one-time credentials stored as hashes. Mark them used atomically and do not log plaintext.
- Barn-to-Barn pairing exchanges scoped credentials: the master gets read probes plus `barn:panel:admin` for `/api/servers/nodes/{id}/proxy/*`; the node gets heartbeat/event write scopes. Never share the global panel `API_TOKEN`.
- Managed nodes accept either `API_TOKEN` or an inbound `master_to_node` panel-admin credential on panel routes (`PanelAuth`). Keep proxy path allowlists tight; do not forward pairing/ingest/agent/node-admin routes.
- Each node probe/ingest endpoint in `router.go` declares a specific scope (`status`, apps, backups, version, heartbeat, or events). Add the narrowest scope and test missing/wrong/revoked credentials.
- Stored remote credentials are encrypted at rest. SSH passwords, private keys, and optional sudo passwords stay only in the in-memory credential store with bounded TTL; never persist or emit them in installation logs.
- Host-key fingerprint confirmation is a required SSH trust step. Do not bypass it to automate installation.

## Change the server or agent workflow

1. Trace API DTO → service → queries/workers/client → frontend consumer. Update all layers and migrations through `barn-db-migration` when persistence changes.
2. Preserve pairing lifecycle: create code/token, validate expiry/hash/target, consume once, exchange only required credentials, establish node mode, and handle disconnect/revocation cleanup.
3. Preserve agent registration lifecycle: a registration token binds to the expected node UID; successful registration writes the scoped node token to a `0600` config; heartbeats and durable event delivery use that credential thereafter.
4. Keep monitored systemd unit names restricted to the existing service-name validation before invoking `systemctl`. Preserve bounded durable outbox behavior for undelivered events.
5. Agent install/update/uninstall jobs affect remote hosts. Maintain in-memory SSH secrets, cancellation and log/status transitions, host-key confirmation, sudo handling, and panel-only deletion (`skip_uninstall`) for unreachable hosts. Agent update replaces binary/restarts service without re-registration or token/config loss.
6. Preserve new and legacy artifacts:

   - `barn-agent` user/binary, `/etc/barn-agent/config.json`, `/var/lib/barn-agent/outbox`, and `barn-agent.service`;
   - `dockpilot-agent` equivalents and legacy API/path fallbacks where currently supported;
   - `linux/amd64` and `linux/arm64` binaries, checksum file, API-image embedding under `/app/agents`, and `BARN_AGENT_DIR` with existing fallback variables.

7. When protocol fields change, design rolling compatibility between old agents and new masters/nodes. Prefer additive fields/defaults; do not require every managed host to upgrade atomically. Update both server and agent docs.

## Verify

```sh
cd backend
gofmt -w <changed-go-files>
go test ./internal/servers ./internal/agent ./internal/metrics ./internal/api
# For protocol, wiring, or shared changes:
go test ./...

cd ..
# When build/packaging changed; writes only dist/agents:
make barn-agent-binaries VERSION=vX.Y.Z
bash -n scripts/<changed-script>.sh
git diff --check
```

Inspect the built checksums and confirm both names and both architectures. If frontend contracts/UI changed, also run TypeScript checking and the appropriate build. Do not run SSH installation, remote uninstall/update, pairing against production, `systemctl`, or release publication as a test unless the user explicitly authorizes the exact targets and external effects.
