# BIOS TUI

The `ctrldot bios` command runs an interactive, full-screen TUI styled like a firmware BIOS screen. It talks to the Ctrl Dot daemon (default `http://127.0.0.1:7777`) and your local config file.

## Running

```bash
ctrldot bios
# Or with a different daemon:
ctrldot --server http://localhost:7777 bios
```

- **q** or **Ctrl+C** — quit
- **↑/↓** — change selection in the left nav
- **Enter** — select (no action yet)
- **r** — refresh data for the current panel
- **Panic panel:** **y** or **e** — enable panic mode; **n** or **d** — disable panic mode

## Panels

- **Status** — Daemon health from `GET /v1/health` (running/not running, version, server URL).
- **Limits** — List of agents from `GET /v1/agents` and per-agent budget/limits from `GET /v1/agents/{id}/limits` (spent/limit and percentage).
- **Rules** — Loaded from config (`~/.ctrldot/config.yaml` or `CTRLDOT_CONFIG`): require resolution, filesystem allow roots, network deny-all, allow domains. Hint to edit with `ctrldot rules edit`.
- **Panic** — Current state from `GET /v1/panic`. Toggle with **y**/**e** (enable) or **n**/**d** (disable); uses `POST /v1/panic/on` and `POST /v1/panic/off`.

## Layout

- **Status bar (top):** “Ctrl Dot BIOS” and “q: quit  r: refresh”
- **Left nav:** Status, Limits, Rules, Panic (brand indigo background, accent for selection)
- **Right panel:** Content for the selected section (live data or errors if daemon is down)
- **Footer:** Nav hint and dimensions

## Styling

- Background: **#3E2C8F** (brand indigo)
- Text: soft white; secondary text: muted lavender
- Selection: accent indigo highlight
- Colours are defined in `internal/ui/bios/style.go`

## Testing without a TTY

```bash
ctrldot bios --print-snapshot
```

Prints a single 80×24 frame of the UI to stdout and exits (useful in CI or non-interactive environments).
