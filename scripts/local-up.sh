#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

# Create .env from example if missing
if [[ ! -f .env ]]; then
  cp .env.example .env
  echo "Created .env from .env.example"
fi

# Load env for this script
set -a
# shellcheck disable=SC1091
source .env
set +a

# Frontend env (Next.js reads .env.local)
write_frontend_env() {
  cat > frontend/.env.local <<EOF
NEXT_PUBLIC_API_URL=${NEXT_PUBLIC_API_URL:-http://localhost:8080}
EOF
}

if [[ ! -f frontend/.env.local ]]; then
  write_frontend_env
  echo "Created frontend/.env.local"
fi

echo "Starting PostgreSQL in Docker..."
if ! docker image inspect barn-postgres:latest >/dev/null 2>&1; then
  echo "Building barn-postgres:latest (Postgres 16 + pgvector)..."
  docker build -t barn-postgres:latest -f docker/postgres/Dockerfile docker/postgres
fi
docker compose up -d postgres

"$ROOT/scripts/wait-for-postgres.sh"
"$ROOT/scripts/migrate.sh"

cat <<'EOF'

Local stack is ready.

  Database:  postgres://dockpilot:dockpilot@localhost:5432/dockpilot

Start the API (terminal 1):
  make backend

Start the UI (terminal 2):
  make frontend

Or run both with:
  make dev-run

Stop database:
  make down

EOF
