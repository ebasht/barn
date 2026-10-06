#!/usr/bin/env bash
# Recover Postgres after an upgrade that attached an empty volume or dropped
# published host ports. Data in Docker volumes is NOT deleted by container rm.
#
#   sudo bash scripts/recover-pg-volume.sh
#   sudo bash scripts/recover-pg-volume.sh --port 18081
#   # legacy DockPilot install:
#   sudo BARN_INSTALL_DIR=/opt/dock-pilot bash scripts/recover-pg-volume.sh --port 18081
#
set -euo pipefail

# Prefer explicit env, then cwd if it looks like an install, then legacy/default paths.
if [[ -n "${BARN_INSTALL_DIR:-${DOCK_PILOT_INSTALL_DIR:-}}" ]]; then
  ROOT="${BARN_INSTALL_DIR:-$DOCK_PILOT_INSTALL_DIR}"
elif [[ -f ./.env ]] && { [[ -f ./docker-compose.full.yml ]] || [[ -f ./docker-compose.barn-full.yml ]] || [[ -f ./docker-compose.dock-pilot.yml ]]; }; then
  ROOT="$(pwd)"
elif [[ -f /opt/dock-pilot/.env ]]; then
  ROOT="/opt/dock-pilot"
else
  ROOT="/opt/barn"
fi
cd "$ROOT"

EXTRA_PORT=""
while [[ $# -gt 0 ]]; do
  case "$1" in
    --port)
      EXTRA_PORT="${2:-}"
      shift 2
      ;;
    --port=*)
      EXTRA_PORT="${1#*=}"
      shift
      ;;
    *)
      echo "Unknown arg: $1" >&2
      exit 1
      ;;
  esac
done

log() { echo "[barn-recover] $*" >&2; }
die() { echo "[barn-recover] ERROR: $*" >&2; exit 1; }

[[ -f .env ]] || die "no .env in $ROOT"
set -a
# shellcheck disable=SC1091
source ./.env
set +a

# Prefer the stack that matches the live/legacy container. Never pick barn-full
# over dock-pilot when dock-pilot-postgres (or its volume) is what is running.
COMPOSE=""
COMPOSE_P=""
LEGACY_PG=""
for v in dock-pilot_dock_pilot_pg dock_pilot_pg dockpilot-postgres-data; do
  if docker volume inspect "$v" >/dev/null 2>&1; then
    LEGACY_PG="$v"
    break
  fi
done

if docker inspect dock-pilot-postgres >/dev/null 2>&1 || [[ -n "$LEGACY_PG" ]]; then
  COMPOSE="docker-compose.full.yml"
  [[ -f "$COMPOSE" ]] || COMPOSE="docker-compose.dock-pilot.yml"
  [[ -f "$COMPOSE" ]] || die "legacy postgres present but docker-compose.full.yml missing in $ROOT"
  COMPOSE_P="dock-pilot"
  log "Legacy DockPilot stack: ${COMPOSE} (-p ${COMPOSE_P})"
else
  COMPOSE="docker-compose.barn-full.yml"
  [[ -f "$COMPOSE" ]] || COMPOSE="docker-compose.barn.yml"
  [[ -f "$COMPOSE" ]] || COMPOSE="docker-compose.full.yml"
  [[ -f "$COMPOSE" ]] || COMPOSE="docker-compose.dock-pilot.yml"
  [[ -f "$COMPOSE" ]] || die "compose file not found in $ROOT"
  log "Barn stack: ${COMPOSE}"
fi

compose() {
  local args=()
  if [[ -n "$COMPOSE_P" ]]; then
    args+=(-p "$COMPOSE_P")
  fi
  args+=(-f "$COMPOSE")
  if [[ -f "$OVERRIDE" ]]; then
    args+=(-f "$OVERRIDE")
  fi
  docker compose "${args[@]}" "$@"
}

PG_USER="${POSTGRES_USER:-barn}"
PG_DB="${POSTGRES_DB:-barn}"
PG_HOST_PORT="${POSTGRES_HOST_PORT:-5433}"
IMAGE="${POSTGRES_IMAGE:-barn-postgres:latest}"

log "Install dir: $ROOT"
log "Listing candidate Postgres volumes..."
CANDIDATES=()
while IFS= read -r v; do
  [[ -n "$v" ]] && CANDIDATES+=("$v")
done < <(docker volume ls --format '{{.Name}}' | grep -E 'pg|postgres|barn|dock' || true)
if ((${#CANDIDATES[@]} == 0)); then
  die "no postgres-like volumes found"
fi
printf '  %s\n' "${CANDIDATES[@]}" >&2

probe_image() {
  if docker image inspect "$IMAGE" >/dev/null 2>&1; then
    echo "$IMAGE"
    return
  fi
  if docker image inspect pgvector/pgvector:pg16 >/dev/null 2>&1; then
    echo "pgvector/pgvector:pg16"
    return
  fi
  if docker image inspect postgres:16-alpine >/dev/null 2>&1; then
    echo "postgres:16-alpine"
    return
  fi
  echo "$IMAGE"
}

PROBE_IMAGE="$(probe_image)"

# Prints: "<score> <volume>"
pick_volume() {
  local vol="$1"
  local tmp="barn-pg-probe-$$"
  docker rm -f "$tmp" >/dev/null 2>&1 || true
  if ! docker run -d --name "$tmp" \
    -v "${vol}:/var/lib/postgresql/data" \
    "$PROBE_IMAGE" >/dev/null 2>&1; then
    return 1
  fi
  local ok=0
  for _ in $(seq 1 30); do
    if docker exec "$tmp" pg_isready >/dev/null 2>&1; then
      ok=1
      break
    fi
    sleep 1
  done
  if [[ "$ok" -ne 1 ]]; then
    docker rm -f "$tmp" >/dev/null 2>&1 || true
    return 1
  fi

  local dbs=0 sites=0 score=0 s u dbname
  dbs="$(docker exec "$tmp" psql -U postgres -d postgres -tAc \
    "SELECT count(*) FROM pg_database WHERE datistemplate = false" 2>/dev/null || echo 0)"
  [[ "$dbs" =~ ^[0-9]+$ ]] || dbs=0

  for u in postgres "$PG_USER" barn dockpilot; do
    [[ -z "$u" ]] && continue
    for dbname in postgres "$PG_DB" barn dockpilot; do
      [[ -z "$dbname" ]] && continue
      s="$(docker exec "$tmp" psql -U "$u" -d "$dbname" -tAc \
        "SELECT count(*) FROM sites" 2>/dev/null || echo "")"
      if [[ "$s" =~ ^[0-9]+$ ]] && ((s > sites)); then
        sites=$s
      fi
    done
  done

  score=$((dbs * 10 + sites))
  log "  volume ${vol}: databases=${dbs} sites=${sites} score=${score}"
  docker rm -f "$tmp" >/dev/null 2>&1 || true
  echo "${score} ${vol}"
}

BEST_SCORE=-1
BEST_VOL=""
# Prefer the volume currently mounted by the live container.
for c in dock-pilot-postgres barn-postgres dockpilot-postgres; do
  if docker inspect "$c" >/dev/null 2>&1; then
    live="$(docker inspect "$c" --format '{{range .Mounts}}{{if eq .Destination "/var/lib/postgresql/data"}}{{.Name}}{{end}}{{end}}' 2>/dev/null || true)"
    if [[ -n "$live" ]]; then
      BEST_VOL="$live"
      BEST_SCORE=999
      log "Using live volume from ${c}: ${BEST_VOL}"
    fi
    break
  fi
done

if [[ -z "$BEST_VOL" ]]; then
  for v in "${CANDIDATES[@]}"; do
    result="$(pick_volume "$v" || true)"
    [[ -z "$result" ]] && continue
    score="${result%% *}"
    vol="${result#* }"
    if [[ "$score" =~ ^[0-9]+$ ]] && ((score > BEST_SCORE)); then
      BEST_SCORE="$score"
      BEST_VOL="$vol"
    fi
  done
fi

[[ -n "$BEST_VOL" ]] || die "could not probe any volume — check docker images and volume list"
if ((BEST_SCORE <= 0)); then
  log "WARN: best volume score is 0 — attaching ${BEST_VOL} anyway"
fi

log "Selected volume: ${BEST_VOL} (score ${BEST_SCORE})"

if [[ -z "$EXTRA_PORT" ]]; then
  # Prefer host_port from panel DB when the container is already up.
  for c in dock-pilot-postgres barn-postgres dockpilot-postgres; do
    if docker inspect "$c" >/dev/null 2>&1; then
      EXTRA_PORT="$(docker exec "$c" psql -U "$PG_USER" -d "$PG_DB" -Atc \
        "SELECT host_port FROM pdb_instances WHERE host_port IS NOT NULL AND host_port > 0 ORDER BY host_port LIMIT 1" 2>/dev/null || true)"
      EXTRA_PORT="$(echo "$EXTRA_PORT" | tr -d '[:space:]')"
      [[ "$EXTRA_PORT" =~ ^[0-9]+$ ]] && break
      EXTRA_PORT="$(docker exec "$c" psql -U postgres -d "$PG_DB" -Atc \
        "SELECT host_port FROM pdb_instances WHERE host_port IS NOT NULL AND host_port > 0 ORDER BY host_port LIMIT 1" 2>/dev/null || true)"
      EXTRA_PORT="$(echo "$EXTRA_PORT" | tr -d '[:space:]')"
      [[ "$EXTRA_PORT" =~ ^[0-9]+$ ]] && break
      EXTRA_PORT=""
    fi
  done
fi
if [[ -z "$EXTRA_PORT" ]]; then
  EXTRA_PORT="18081"
  log "Publishing managed app port :${EXTRA_PORT} (override with --port)"
else
  log "Publishing managed app port :${EXTRA_PORT}"
fi

OVERRIDE="${ROOT}/docker-compose.barn-pgdata.yml"
PG_COMPOSE_VOL="barn_pg"
if grep -qE 'dock_pilot_pg(:|/var/lib/postgresql/data)' "$COMPOSE" 2>/dev/null; then
  PG_COMPOSE_VOL="dock_pilot_pg"
fi
cat >"$OVERRIDE" <<EOF
# Generated by recover-pg-volume.sh — remap compose volume
services:
  postgres:
    ports:
      - "0.0.0.0:${EXTRA_PORT}:5432"
volumes:
  ${PG_COMPOSE_VOL}:
    external: true
    name: ${BEST_VOL}
EOF

log "Stopping current postgres/api..."
compose stop postgres api 2>/dev/null || true
docker rm -f barn-postgres dock-pilot-postgres dockpilot-postgres barn-pg-probe-$$ 2>/dev/null || true

log "Starting postgres on recovered volume..."
compose up -d --force-recreate --no-deps postgres

log "Waiting for postgres..."
ready=0
for _ in $(seq 1 45); do
  if compose exec -T postgres pg_isready >/dev/null 2>&1; then
    ready=1
    break
  fi
  sleep 1
done
[[ "$ready" -eq 1 ]] || die "postgres did not become ready"

log "Databases:"
compose exec -T postgres \
  psql -U "$PG_USER" -d postgres -c \
  "SELECT datname FROM pg_database WHERE datistemplate = false ORDER BY 1" 2>/dev/null || \
compose exec -T postgres \
  psql -U postgres -d postgres -c \
  "SELECT datname FROM pg_database WHERE datistemplate = false ORDER BY 1" || true

log "Running migrations..."
compose run --rm -T migrate || \
  log "WARN: migrate failed — check POSTGRES_USER/DB in .env match the volume"

log "Recreating api + frontend..."
compose up -d --force-recreate api frontend

# Pin port into .env so base compose keeps publishing it on future upgrades.
if grep -qE '^MANAGED_PG_HOST_PORT=' .env 2>/dev/null; then
  tmp="$(mktemp)"
  awk -v v="$EXTRA_PORT" '/^MANAGED_PG_HOST_PORT=/{print "MANAGED_PG_HOST_PORT="v; next} {print}' .env >"$tmp"
  mv "$tmp" .env
else
  printf '\nMANAGED_PG_HOST_PORT=%s\n' "$EXTRA_PORT" >> .env
fi
log "Force-publishing managed port via docker run..."
PUBLISH="${ROOT}/scripts/publish-managed-pg-port.sh"
[[ -f "$PUBLISH" ]] || die "publish-managed-pg-port.sh missing"
env BARN_INSTALL_DIR="$ROOT" DOCK_PILOT_INSTALL_DIR="$ROOT" bash "$PUBLISH" --port "$EXTRA_PORT"

log "Done. Volume ${BEST_VOL} attached; override: $OVERRIDE"
log "Panel: 127.0.0.1:${PG_HOST_PORT}  Apps: 0.0.0.0:${EXTRA_PORT}"
docker port dock-pilot-postgres 2>/dev/null || docker port barn-postgres 2>/dev/null || true
