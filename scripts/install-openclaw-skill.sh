#!/usr/bin/env bash
# Copy Ctrl Dot skill to OpenClaw skills directory.
# Run from repo root.

set -e
cd "$(dirname "$0")/.."

SKILLS_DIR="${OPENCLAW_SKILLS_DIR:-$HOME/.openclaw/skills}"
mkdir -p "$SKILLS_DIR"
cp -r adapters/openclaw/skills/ctrldot-guard "$SKILLS_DIR/"
echo "Installed skill to $SKILLS_DIR/ctrldot-guard"
echo "Restart OpenClaw to load the skill."
