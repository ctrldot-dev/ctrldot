# Ctrl Dot Documentation Index

This folder is the source of truth for Ctrl Dot documentation. The `ctrldot-docs` MCP server indexes these Markdown files and exposes **List Docs**, **Fetch Doc**, and **Search Docs** tools for agents and IDEs.

## Overview

- Use **SETUP_GUIDE** to install and run Ctrl Dot with adapters (CrewAI, OpenClaw).
- Use **CONFIG** to understand and edit `~/.ctrldot/config.yaml`.
- Use **AGENT_EXPERIENCE** to integrate agents and interpret Ctrl Dot decisions.
- Use **SHOWCASE**, **DECISION_LEDGER_DEMO_IMPLEMENTATION_PLAN**, and **BIOS_UI** for deeper or experimental flows.

## Docs catalogue

| Doc | Purpose |
|-----|---------|
| `SETUP_GUIDE.md` | End‑to‑end installation and integration guide (daemon, CrewAI, OpenClaw, smoke tests). |
| `CONFIG.md` | Configuration reference for `~/.ctrldot/config.yaml` and related env vars. |
| `AGENT_EXPERIENCE.md` | How agents detect Ctrl Dot, interpret DENY/STOP/THROTTLE responses, and guide users. |
| `TROUBLESHOOTING.md` | Common problems (daemon, database, adapters) and how to fix them. |
| `SHOWCASE.md` | Example flows demonstrating Ctrl Dot decisions and capabilities. |
| `BIOS_UI.md` | Notes on the BIOS‑style TUI (`ctrldot bios`), layout, and behaviour. |
| `DECISION_LEDGER_DEMO_IMPLEMENTATION_PLAN.md` | Design/implementation plan for the decision ledger demo and related API surface. |

## Using these docs via MCP

When `ctrldot-docs` is configured as an MCP server:

1. Call **Search Docs** with a short query (e.g. `"panic mode"`, `"daily_budget_gbp"`).
2. From the results, use **Fetch Doc** on the best match to get the full Markdown (plus headings and anchors).
3. Ask the model to answer questions **based only on the fetched text**, citing:
   - `path` (e.g. `docs/SETUP_GUIDE.md`)
   - `heading.anchor` for deep links within the page.

