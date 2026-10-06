---
name: barn-install-upgrade
description: >
  Change or review Barn VPS installation, upgrade, environment, Docker Compose,
  nginx, TLS, filesystem, service, and legacy DockPilot compatibility workflows.
---

# Barn Install and Upgrade

Use this workflow for `scripts/install*.sh`, `barn-up*.sh`, `barn-upgrade.sh`, Compose install variants, panel nginx templates, systemd units, and legacy migration behavior.

## Model the supported paths first

Read `AGENTS.md`, `scripts/AGENTS.md`, the relevant README install/upgrade sections, and every script/template/Compose file touched by the change. Trace both new and existing installations:

- New Barn installs default to `/opt/barn`, Barn image/container/volume names, `docker-compose.barn-full.yml`, and `.env.barn.example`.
- Legacy installations may live at `/opt/dock-pilot` and use DockPilot image, container, volume, Compose, env, work-directory, agent, and service names. `dock-pilot-*` scripts are compatibility wrappers, not dead code.
- `install.sh` bootstraps a release and may generate secrets, packages, `.env`, nginx/TLS configuration, volumes, and services. Re-running install is not the supported image upgrade path.
- `barn-upgrade.sh` preserves `.env` and data, downloads Barn bundles with a DockPilot fallback, loads images, selects the real PostgreSQL volume, migrates, recreates services, and may reconfigure nginx. It deliberately re-execs from the extracted copy so it does not overwrite the running script.
- Domain installs use nginx and normally certbot; `--skip-cert` is HTTP-only. No-domain installs use `PANEL_HTTP_PORT` (default 8888), IP-oriented nginx, and matching CORS origins.

Search every compatibility name before changing one:

```sh
rg -n 'barn|dock-pilot|dockpilot|BARN_|DOCK_PILOT_|/opt/|/var/lib/|compose|container_name|volume' \
  scripts install docker-compose*.yml .env*.example README.md backend frontend
```

## Implement safely

- Preserve `BARN_*` → `DOCK_PILOT_*` fallbacks and legacy wrappers unless removal is explicitly scoped and migration evidence exists.
- Keep install and upgrade semantics separate. An upgrade must not rotate API/DB/encryption secrets, replace a populated volume with an empty one, or assume `POSTGRES_PASSWORD` changes an existing PGDATA.
- Treat `--reset-db`, reset/recover/restore helpers, volume selection/removal, credential rewrites, and service replacement as destructive. Require explicit intent and exact targets.
- Keep the release bundle topology aligned with install/upgrade lookup order: Barn and legacy bundle/image names, required Compose variants, `scripts/`, `install/`, env examples, and `VERSION`.
- When changing environment variables, update the appropriate `.env*.example`, config loader, Compose service, installer/upgrade preservation logic, and docs. Never print or inspect a working `.env`; use examples.
- When changing panel proxying, update both domain and IP templates and retain `/mcp`, `/api`, SSE/log streaming behavior, body limits, Host/Origin/CORS assumptions, `nginx -t` before reload, and certificate behavior.
- When changing agent units or paths, update both `install/barn-agent.service` and the legacy service where compatibility is affected.
- Keep scripts Bash with `set -euo pipefail`; quote expansions and validate option arity, paths, downloads, temp files, and cleanup. Do not weaken failure handling around migration or data selection without documenting the recovery behavior.

## Safe validation

Static checks are the default:

```sh
bash -n scripts/<changed-script>.sh
# Include wrappers/helpers that source or exec it:
bash -n scripts/install.sh scripts/install-lib.sh scripts/barn-upgrade.sh \
  scripts/barn-up.sh scripts/dock-pilot-upgrade.sh scripts/dock-pilot-up.sh

# Render only with example/synthetic environment values; never log real secrets.
docker compose --env-file .env.barn.example -f docker-compose.barn-full.yml config >/dev/null
docker compose --env-file .env.dock-pilot.example -f docker-compose.full.yml config >/dev/null
git diff --check
```

Also inspect option help and branching for: fresh Barn, existing Barn, legacy DockPilot, domain+TLS, domain+`--skip-cert`, no-domain, offline/`--from-dir`, and supported amd64/arm64 artifacts where relevant.

Do not run install, upgrade, migrate, reset, recovery, restore, certbot, nginx reload, package installation, or production Compose changes merely as validation. If an end-to-end test is explicitly requested, use an isolated disposable host/VM and state exactly which destructive paths are authorized.
