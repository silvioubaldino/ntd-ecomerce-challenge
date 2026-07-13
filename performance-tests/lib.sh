#!/usr/bin/env bash
# Shared functions from performance-tests/README.md guide.
# Usage: source performance-tests/lib.sh   (once, at the start of your terminal session)

export DSN="postgres://ntd:ntd@localhost:5432/ntd_ecomerce?sslmode=disable"
export PGURL="postgresql://ntd:ntd@localhost:5432/ntd_ecomerce"

_LIB_SOURCE="${BASH_SOURCE[0]:-$0}"
SCRIPT_DIR="$(cd "$(dirname "$_LIB_SOURCE")" && pwd)"
COMPOSE_FILE="$SCRIPT_DIR/docker-compose.perf.yml"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

db_fresh() {   # bring database up from scratch (delete volume)
  docker compose -f "$COMPOSE_FILE" down -v
  docker compose -f "$COMPOSE_FILE" up -d
  until docker compose -f "$COMPOSE_FILE" exec -T db pg_isready -U ntd -d ntd_ecomerce >/dev/null 2>&1; do sleep 1; done
}

db_down() { docker compose -f "$COMPOSE_FILE" down -v; }

api_up() {   # $1 = commit hash
  local hash=$1
  local dir="/tmp/perf-$hash"
  git -C "$REPO_ROOT" worktree add -f "$dir" "$hash"
  ( cd "$dir/ntd-ecomerce-api" \
    && DATABASE_URL="$DSN" API_PORT=8080 go run ./cmd/api ) &
  echo $! > /tmp/perf-api.pid
  until curl -sf localhost:8080/products >/dev/null 2>&1; do sleep 1; done
}

api_down() {
  # `go run` starts a child binary that survives the kill of the parent process ($!);
  # kill whoever is actually listening on the port to avoid leaving a zombie holding 8080.
  kill "$(cat /tmp/perf-api.pid 2>/dev/null)" 2>/dev/null || true
  lsof -tiTCP:8080 -sTCP:LISTEN 2>/dev/null | xargs -r kill 2>/dev/null || true
  sleep 1
}

worktrees_clean() { git -C "$REPO_ROOT" worktree prune; }

seed_db() {   # $1 = number of rows (default 500000). Requires psql installed.
  local n="${1:-500000}"
  sed "s/generate_series(1, [0-9]*)/generate_series(1, $n)/" "$SCRIPT_DIR/seed.sql" \
    | psql "$PGURL"
}
