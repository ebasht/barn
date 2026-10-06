---
name: barn-api-change
description: >
  Implement or review Barn HTTP API endpoint, contract, authentication, authorization,
  error, frontend-client, and UI-consumer changes across the Go API and Next.js frontend.
---

# Barn API Change

Use this workflow for a new or changed `/api` endpoint or any Go JSON contract consumed by the frontend. For MCP protocol changes, use `barn-mcp-change`; for schema work, also use `barn-db-migration`.

## Establish the contract and its consumers

1. Read `AGENTS.md`, `backend/AGENTS.md`, and, when the browser contract or UI changes, `frontend/AGENTS.md`.
2. Locate the route in `backend/internal/api/router.go`, then follow its handler in `backend/internal/api/*_handler.go` into `backend/internal/<domain>/service.go`, `dto.go`, errors, workers, and queries.
3. Search by all contract identifiers, not only the URL:

   ```sh
   rg -n 'route-fragment|HandlerName|RequestType|ResponseType|json_field' backend frontend docs
   rg -n 'api\.(methodName)|json_field|TypeScriptType' frontend/app frontend/components frontend/lib
   ```

4. Inspect `frontend/lib/api.ts`, `frontend/lib/types.ts`, and `frontend/lib/normalize.ts`. Search pages and components that consume the API method/type. Include exports/imports, tests, SSE callers, and documentation. Do not infer JSON names from Go field names; use explicit tags and the actual TypeScript shape.
5. Classify the route's trust boundary before editing:

   - panel routes are inside the `BearerTokenAuth` group;
   - `/health` and QR exchange are deliberately public;
   - most server/node endpoints use scoped server-token middleware per route;
   - pairing and agent-registration entrypoints sit outside panel auth but must validate their short-lived one-time credentials in the service;
   - SSE panel endpoints may accept `?token=` because browser `EventSource` cannot set headers;
   - MCP settings endpoints use panel auth, while `/mcp` has separate MCP bearer authentication.

Do not move a route between these boundaries merely for convenience. If adding a cross-origin header or method, update the CORS configuration deliberately.

## Implement end to end

- Register the route in `router.go` with the correct chi group and middleware. Wire a new handler/service through `Handlers`, `Mount`, and `backend/cmd/server/main.go` only when the existing domain wiring cannot be reused.
- Keep HTTP parsing, URL-parameter validation, and status selection in the handler. Put business rules and persistence in the domain service. Follow adjacent DTO and mapper patterns.
- Validate request bodies and UUIDs explicitly. Use the established domain errors and `writeError`; extend `backend/internal/api/response.go` when a new stable error category needs an HTTP mapping. Return JSON through `writeJSON`, `204` only for intentionally empty success, and non-nil empty slices when clients expect arrays.
- Preserve existing response fields unless the task explicitly authorizes a breaking change. For renames, consider a compatibility field or frontend normalizer before removing the old spelling.
- Update `frontend/lib/api.ts` and `frontend/lib/types.ts` together. Add normalization in `normalize.ts` when deployed backends may return legacy or nullable shapes. Use the shared API base and auth/error handling rather than raw `fetch` in components.
- If UI behavior or visible copy changes, update the relevant component/page and both `frontend/lib/i18n/messages/en.ts` and `ru.ts`. Check mobile behavior for changed controls.
- Add focused handler/service tests near the changed package. Cover authorization, malformed input, domain-error mapping, success status/body, and compatibility behavior relevant to the change.

## Verify

Run from the indicated directory, scaling from focused to broad:

```sh
cd backend
gofmt -w <changed-go-files>
go test ./internal/<changed-domain> ./internal/api
# Use for shared contracts, router/wiring, or several domains:
go test ./...

cd ../frontend
npx --no-install tsc --noEmit --incremental false
# If a touched utility has a node:test file:
node --experimental-strip-types --test lib/env-file.test.ts lib/site-import.test.ts
# For routes, shared shell, or build behavior:
npm run build

cd ..
git diff --check
```

Do not claim `npm run lint` as a gate: the repository has no working ESLint configuration. If a schema/query changed, complete `barn-db-migration` verification too. Summarize the changed contract, every updated consumer, auth boundary, compatibility decision, and commands actually run.
