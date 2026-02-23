#!/usr/bin/env bash
# Build Ctrl Dot OpenClaw plugin.
# Run from repo root.

set -e
cd "$(dirname "$0")/.."

echo "=== Building Ctrl Dot OpenClaw plugin ==="
cd adapters/openclaw/plugin
npm install
npm run build
echo "Done. Plugin output: $(pwd)/dist/"
echo "Copy adapters/openclaw/examples/openclaw-home-config.example.json to ~/.openclaw/openclaw.json and set ABSOLUTE_PATH_TO_REPO."
