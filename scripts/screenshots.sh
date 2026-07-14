#!/usr/bin/env bash
# Generates the README screenshots: starts a throwaway demo server + a real agent
# on this machine, then drives a headless browser to capture the images (English
# and German) into docs/screenshots/.
#
# Requirements: Go and Node.js. Playwright + its Chromium are installed on demand.
#   ./scripts/screenshots.sh
set -euo pipefail
cd "$(dirname "$0")/.."

DD="$(mktemp -d)"
OUT="docs/screenshots"; mkdir -p "$OUT"
PORT="${PORT:-18080}"
# On exit: stop demo processes, remove temp dir, and restore the committed
# placeholder web/dist/index.html (the build step overwrote it with hashed asset
# references that must not be committed).
trap 'kill $(jobs -p) 2>/dev/null || true; rm -rf "$DD"; git checkout -- web/dist/index.html 2>/dev/null || true' EXIT

echo "→ installing web dependencies + Playwright/Chromium…"
( cd web && npm install >/dev/null 2>&1 )
( cd web && npm ls playwright >/dev/null 2>&1 || npm install --no-save playwright >/dev/null 2>&1 )
( cd web && npx playwright install chromium >/dev/null 2>&1 || true )

echo "→ building web UI (embedded into the server)…"
( cd web && npm run build >/dev/null 2>&1 )

echo "→ building server + agent…"
CGO_ENABLED=0 go build -o "$DD/server" ./cmd/server
CGO_ENABLED=0 go build -o "$DD/agent"  ./cmd/agent

cat > "$DD/agent.yaml" <<YML
server_url: http://127.0.0.1:$PORT
enrollment_token: demotoken
state_path: $DD/state.json
disable_public_ip: true
disable_update_check: true
disable_auto_update: true
YML

echo "→ starting demo server + agent…"
ROSTER_DB="sqlite://$DD/demo.db" ROSTER_ADDR="127.0.0.1:$PORT" \
  ROSTER_SEED_ADMIN_USER=admin ROSTER_SEED_ADMIN_PASSWORD=demo1234 \
  ROSTER_REQUIRE_2FA=false ROSTER_SEED_ENROLL_TOKEN=demotoken ROSTER_CHECKIN_INTERVAL=5s \
  "$DD/server" run &
sleep 3
"$DD/agent" -config "$DD/agent.yaml" run &
sleep 12  # let the agent enroll + check in a few times

echo "→ capturing screenshots into $OUT …"
( cd web && node ../scripts/capture.mjs "http://127.0.0.1:$PORT" "../$OUT" )
echo "✓ done → $OUT"
echo "  Review the PNGs, then embed them in README.md and commit."
