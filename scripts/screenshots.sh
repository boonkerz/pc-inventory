#!/usr/bin/env bash
# Generates the README screenshots: starts a local demo server + a real agent on
# this machine, then drives a headless browser to capture the images.
#
# Requirements: Go, Node.js, and Playwright's Chromium:
#   cd web && npm install && npx playwright install chromium
set -euo pipefail
cd "$(dirname "$0")/.."

DD="$(mktemp -d)"
OUT="docs/screenshots"; mkdir -p "$OUT"
PORT="${PORT:-18080}"
trap 'kill $(jobs -p) 2>/dev/null || true; rm -rf "$DD"' EXIT

echo "building server + agent…"
CGO_ENABLED=0 go build -o "$DD/server" ./cmd/server
CGO_ENABLED=0 go build -o "$DD/agent" ./cmd/server >/dev/null 2>&1 || CGO_ENABLED=0 go build -o "$DD/agent" ./cmd/agent

cat > "$DD/agent.yaml" <<YML
server_url: http://127.0.0.1:$PORT
enrollment_token: demotoken
state_path: $DD/state.json
disable_public_ip: true
disable_update_check: true
disable_auto_update: true
YML

echo "starting demo server + agent…"
PCINV_DB="sqlite://$DD/demo.db" PCINV_ADDR="127.0.0.1:$PORT" \
  PCINV_SEED_ADMIN_USER=admin PCINV_SEED_ADMIN_PASSWORD=demo1234 \
  PCINV_REQUIRE_2FA=false PCINV_SEED_ENROLL_TOKEN=demotoken PCINV_CHECKIN_INTERVAL=5s \
  "$DD/server" run &
sleep 3
"$DD/agent" -config "$DD/agent.yaml" run &
sleep 12  # let the agent enroll + check in a few times

echo "capturing screenshots into $OUT …"
( cd web && node ../scripts/capture.mjs "http://127.0.0.1:$PORT" "../$OUT" )
echo "done -> $OUT"
