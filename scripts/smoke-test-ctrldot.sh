#!/usr/bin/env bash
# Smoke test: Ctrl Dot with SQLite (no Postgres). Build, start daemon, status, doctor, register, propose, events, stop.
# Run from repo root.

set -e
cd "$(dirname "$0")/.."

export CTRLDOT_RUNTIME_STORE=sqlite
export CTRLDOT_BUNDLE_DIR="${CTRLDOT_BUNDLE_DIR:-/tmp/ctrldot-smoke-bundles}"
rm -rf "$CTRLDOT_BUNDLE_DIR"
mkdir -p "$CTRLDOT_BUNDLE_DIR"

echo "=== Build ==="
go build -o bin/ctrldot ./cmd/ctrldot
go build -o bin/ctrldotd ./cmd/ctrldotd
go build -o bin/ctrldot-mcp ./cmd/ctrldot-mcp

echo "=== Doctor (before daemon) ==="
./bin/ctrldot doctor

echo "=== Start daemon (SQLite + bundle sink) ==="
CTRLDOT_LEDGER_SINK=bundle ./bin/ctrldotd &
DPID=$!
cleanup() { set +e; kill -TERM $DPID 2>/dev/null; wait $DPID 2>/dev/null; set -e; }
trap cleanup EXIT
sleep 2

echo "=== Status ==="
./bin/ctrldot status

echo "=== Capabilities (agent discovery) ==="
CAPS=$(curl -s http://127.0.0.1:7777/v1/capabilities)
echo "$CAPS" | grep -q '"runtime_store"' && echo "$CAPS" | grep -q '"ledger_sink"' && echo "$CAPS" | grep -q '"panic"' && echo "$CAPS" | grep -q '"features"'
echo "$CAPS" | grep -q 'auto_bundles' || true
if [ -f ./bin/ctrldot-mcp ]; then
  ./bin/ctrldot-mcp capabilities | grep -q ctrldot
fi

echo "=== Register agent ==="
curl -s -X POST http://127.0.0.1:7777/v1/agents/register \
  -H "Content-Type: application/json" \
  -d '{"agent_id":"smoke-test","display_name":"Smoke Test","default_mode":"normal"}' | grep -q agent_id

echo "=== Start session ==="
SESS=$(curl -s -X POST http://127.0.0.1:7777/v1/sessions/start \
  -H "Content-Type: application/json" \
  -d '{"agent_id":"smoke-test"}' | sed -n 's/.*"session_id":"\([^"]*\)".*/\1/p')
test -n "$SESS"

echo "=== Propose action ==="
curl -s -X POST http://127.0.0.1:7777/v1/actions/propose \
  -H "Content-Type: application/json" \
  -d '{"agent_id":"smoke-test","session_id":"'"$SESS"'","intent":{"title":"t"},"action":{"type":"tool.call","target":{},"inputs":{}},"cost":{"currency":"GBP","estimated_gbp":0.01,"estimated_tokens":1,"model":""},"context":{"tool":"test","hash":"h"}}' | grep -q decision

echo "=== Events tail ==="
./bin/ctrldot events tail --n 5 | grep -q Events

echo "=== Panic + autobundle ==="
./bin/ctrldot panic on
./bin/ctrldot panic status | grep -q "ON"
./bin/ctrldot autobundle status | grep -q "Auto-bundles"
# Trigger a DENY (e.g. missing resolution) to create an autobundle and get recommendation
DENY_RESP=$(curl -s -X POST http://127.0.0.1:7777/v1/actions/propose \
  -H "Content-Type: application/json" \
  -d '{"agent_id":"smoke-test","session_id":"'"$SESS"'","intent":{"title":"t"},"action":{"type":"git.push","target":{},"inputs":{}},"cost":{"currency":"GBP","estimated_gbp":0.01,"estimated_tokens":1,"model":""},"context":{"hash":"h2"}}')
echo "$DENY_RESP" | grep -q decision
echo "$DENY_RESP" | grep -q '"recommendation"'
echo "$DENY_RESP" | grep -q 'next_steps'
./bin/ctrldot panic off

echo "=== Stop daemon (flush bundle) ==="
kill -TERM $DPID 2>/dev/null || true
wait $DPID 2>/dev/null || true
sleep 1
# Ignore daemon exit code (e.g. http: Server closed)
true

echo "=== Bundle ls ==="
CTRLDOT_BUNDLE_DIR="$CTRLDOT_BUNDLE_DIR" ./bin/ctrldot bundle ls
test -n "$(ls -A "$CTRLDOT_BUNDLE_DIR" 2>/dev/null)"

echo "=== Bundle verify ==="
BUNDLE_DIR=$(ls -d "$CTRLDOT_BUNDLE_DIR"/bundle_* 2>/dev/null | head -1)
if [ -n "$BUNDLE_DIR" ]; then
  ./bin/ctrldot bundle verify "$BUNDLE_DIR"
  test -f "$BUNDLE_DIR/README.md" && echo "README.md present"
  grep -q "ctrldot" "$BUNDLE_DIR/README.md" && echo "README.md contains ctrldot command"
fi

echo "=== Smoke test passed ==="
exit 0
