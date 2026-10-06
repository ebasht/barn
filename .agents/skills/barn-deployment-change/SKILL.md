---
name: barn-deployment-change
description: >
  Implement, test, or diagnose Barn site deployment orchestration across Docker,
  Git, nginx, TLS, health, status/log streaming, and stub versus real modes.
---

# Barn Deployment Change

Use this workflow for site deployment, container lifecycle, deployment state/logs, Docker adapters, nginx, certbot, and health behavior. Installer-level panel nginx/Compose work belongs in `barn-install-upgrade`.

## Map the boundary being changed

Read `AGENTS.md` and `backend/AGENTS.md`. Follow the relevant path through:

- `backend/internal/api/deployments_handler.go` and site container/log handlers;
- `backend/internal/deployments/service.go`, `worker.go`, DTOs, mapper, errors, and queries;
- `backend/internal/sites` for site/container/health behavior;
- `backend/internal/docker`, `nginx`, and `ssl` interfaces plus stub/real implementations and factories;
- `backend/internal/config/config.go` and construction in `backend/cmd/server/main.go`;
- frontend deployment/log/health consumers when contracts change.

The service creates a `pending` deployment and enqueues it. The worker records logs/status while it clones, builds, allocates a port when needed, runs the container, writes/tests/reloads nginx, issues TLS, and reapplies HTTPS configuration. Telegram bots and host-network sites intentionally take different branches.

## Implement without collapsing abstractions

- Add capabilities to the `docker.Client`, `nginx.Manager`, or `ssl.Manager` boundary when orchestration needs them; update both real and stub implementations and factory wiring. Do not type-assert a concrete real implementation except for an explicitly optional capability already modeled that way.
- Keep `DEPLOY_MODE=stub|real` selection in factories. Stub mode must avoid Docker socket, host nginx, certificate, filesystem, and network side effects while still producing useful deterministic logs/results for development and tests.
- Keep orchestration and persisted state in `deployments`; reusable container mechanics in `docker`; site-specific lifecycle/health in `sites`; proxy rendering/application in `nginx`; certificate work in `ssl`.
- Preserve Barn/DockPilot compatibility in image/container/volume naming helpers and existing stored site configurations.
- Maintain the status lifecycle and append actionable logs without tokens, decrypted secrets, private keys, environment values, or other credentials. A failure must finish the deployment as failed and retain the useful step error.
- Test nginx configuration before reload. Treat certificate issuance, container replacement, pruning, delete, and real deploys as host-changing operations.
- Keep log APIs bounded and ordered. SSE streaming must retain authentication, event IDs, flush behavior, disconnect handling, and terminal status semantics.

## Safe implementation verification

Prefer unit tests against interfaces/stubs and pure helpers:

```sh
cd backend
gofmt -w <changed-go-files>
go test ./internal/deployments ./internal/docker ./internal/nginx ./internal/ssl ./internal/sites
# Include API/config/wiring or cross-package changes:
go test ./...
cd ..
git diff --check
```

For a local behavioral check, use the repository's normal development configuration with `DEPLOY_MODE=stub`; verify queued → running → terminal state and logs without assuming it proves real Docker/nginx/certbot behavior. Do not switch to `real`, trigger repeated real deployments, reload host nginx, request certificates, or prune containers just to validate code.

## Diagnose failures

1. Capture the deployment ID, persisted status/message, and bounded deployment logs; distinguish enqueue failure from a worker-step failure.
2. Identify the first failed step and inspect the responsible boundary. For container health/log diagnosis, resolve all compatible container names through existing helpers rather than guessing a single Barn name.
3. Check configuration mode and paths (`DEPLOY_MODE`, work dir, Docker host, host root, nginx sites dirs, certbot email) using examples/config code without exposing working secrets.
4. For nginx, render/inspect the intended config and run a non-mutating config test where authorized before any reload. For TLS, separate DNS/webroot/port failures from config generation.
5. A timeout or lost API response does not prove the deploy was not queued. Check deployment history/status before retrying; each retry creates another deployment.

Report which mode was tested, which external effects were intentionally not exercised, and any real-host validation still required.
