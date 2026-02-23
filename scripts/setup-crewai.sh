#!/usr/bin/env bash
# CrewAI adapter setup: create venv, install deps, run example.
# Run from repo root. Ctrl Dot daemon must be running (see docs/SETUP_GUIDE.md).
# No Postgres required: start with ./bin/ctrldot daemon start (SQLite default).

set -e
cd "$(dirname "$0")/.."

echo "=== CrewAI adapter setup ==="
python3 -m venv .venv
source .venv/bin/activate
pip install --upgrade pip
pip install crewai
pip install -e adapters/crewai
echo ""
echo "=== Running crew_minimal.py (ensure Ctrl Dot daemon is running on :7777) ==="
echo "    Start daemon with: ./bin/ctrldot daemon start (no Postgres needed)"
python adapters/crewai/examples/crew_minimal.py
