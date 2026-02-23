# Ctrl Dot Adapters

This directory contains adapters for integrating Ctrl Dot with popular AI frameworks.

## Available Adapters

### CrewAI (Python)

**Location**: `crewai/`

Python adapter for guarding CrewAI LLM calls and tool execution.

- **Client**: HTTP client for Ctrl Dot API
- **LLM Wrapper**: Guards all LLM calls
- **Tool Wrapper**: Guards all tool executions
- **Throttle Handling**: Automatically reduces max_tokens/temperature on THROTTLE

See `crewai/README.md` for installation and usage.

### OpenClaw (TypeScript)

**Location**: `openclaw/`

OpenClaw plugin that replaces risky built-in tools with Ctrl Dot guarded versions.

- **Plugin**: TypeScript plugin for OpenClaw
- **Tools**: `ctrldot_exec`, `ctrldot_web_fetch`, `ctrldot_propose`
- **Skill**: Instructs agents to use guarded tools
- **Config**: Example OpenClaw configuration

See `openclaw/README.md` for setup and usage.

## Quick Start

### CrewAI

```bash
cd adapters/crewai
pip install -e .
pip install crewai
python examples/crew_minimal.py
```

### OpenClaw

```bash
cd adapters/openclaw/plugin
npm install
npm run build
# Then configure OpenClaw to load the plugin
```

## Architecture

Both adapters follow the same pattern:

1. **Client Layer**: HTTP client wrapping Ctrl Dot API
2. **Guard Layer**: Wrappers that propose actions before execution
3. **Decision Handling**: 
   - ALLOW/WARN: Proceed
   - THROTTLE: Proceed with degraded constraints
   - DENY/STOP: Raise exception/error

## Contributing

To add a new adapter:

1. Create a new directory under `adapters/`
2. Implement client, guards, and examples
3. Add README with installation and usage
4. Update this README
