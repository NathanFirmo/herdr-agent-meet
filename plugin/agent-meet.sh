#!/bin/sh
# Runs Agent Meet for the Herdr plugin with the repository's compose.yaml and
# the published image, so installing the plugin needs Docker and nothing else.
set -eu

ROOT=${HERDR_PLUGIN_ROOT:-$(cd "$(dirname "$0")/.." && pwd)}
CONFIG_DIR=${HERDR_PLUGIN_CONFIG_DIR:-$ROOT}
ENV_FILE=$CONFIG_DIR/.env
SOCKET=${HERDR_SOCKET_PATH:-$HOME/.config/herdr/herdr.sock}

if [ ! -f "$ENV_FILE" ]; then
  mkdir -p "$CONFIG_DIR"
  cp "$ROOT/.env.example" "$ENV_FILE"
  echo "Created $ENV_FILE; edit it to configure Agent Meet."
fi

setting() {
  sed -n "s/^$1=//p" "$ENV_FILE" | tail -n 1
}

PORT=$(setting PORT)
PORT=${PORT:-8080}
UPLOAD_DIR=$(setting UPLOAD_DIR)
UPLOAD_DIR=${UPLOAD_DIR:-/tmp/agent-meet}

compose() {
  # The Herdr directory is mounted rather than the socket file, so a restarted server stays reachable.
  env UID="$(id -u)" GID="$(id -g)" HERDR_DIR="$(dirname "$SOCKET")" \
    docker compose --project-name herdr-agent-meet --project-directory "$ROOT" \
    --env-file "$ENV_FILE" -f "$ROOT/compose.yaml" "$@"
}

running() {
  [ -n "$(compose ps --status running --quiet app 2>/dev/null)" ]
}

start() {
  mkdir -p "$UPLOAD_DIR"
  compose up -d --no-build --pull missing
}

open_url() {
  url=http://localhost:$PORT
  for _ in $(seq 1 20); do
    curl -fs -o /dev/null "$url" && break
    sleep 1
  done
  case "$(uname -s)" in
    Darwin) open "$url" ;;
    *) xdg-open "$url" >/dev/null 2>&1 & ;;
  esac
}

case "${1:-}" in
  start) start ;;
  stop) compose down ;;
  update)
    compose pull
    start
    ;;
  open)
    running || start
    open_url
    ;;
  logs) compose logs -f --tail 100 app ;;
  status) compose ps ;;
  *)
    echo "usage: $0 start|stop|update|open|logs|status" >&2
    exit 2
    ;;
esac
