# Dot CLI - Usage Guide

Dot is a git-like command-line interface for the Futurematic Kernel.

## Installation

### Build from Source

```bash
# Build the dot CLI
go build -o bin/dot ./cmd/dot

# Or add to PATH
export PATH=$PATH:$(pwd)/bin
```

## Quick Start

### 1. Start the Kernel Server

First, make sure the kernel server is running:

```bash
# In the kernel directory
make dev
```

The server will start on `http://localhost:8080` by default.

### 2. Configure Dot CLI

```bash
# Set the server URL (if different from default)
dot config set server http://localhost:8080

# Set your actor ID
dot config set actor_id user:alice

# Set your namespace
dot config set namespace_id ProductTree:/MyProject

# Or use the convenience command
dot use ProductTree:/MyProject

# View your configuration
dot whereami
```

### 3. Check Status

```bash
# Check if kernel is reachable
dot status
```

## Commands

### Connectivity & Configuration

#### `dot status`
Check kernel server health and display current configuration.

```bash
dot status
# Output: server=http://localhost:8080 actor=user:alice namespace=ProductTree:/MyProject ok=true

dot status --json
# Output: {"ok":true,"server":"http://localhost:8080","actor_id":"user:alice","namespace_id":"ProductTree:/MyProject"}
```

#### `dot config get <key>`
Get a configuration value.

```bash
dot config get server
dot config get actor_id
dot config get namespace_id
dot config get capabilities
```

#### `dot config set <key> <value>`
Set a configuration value.

```bash
dot config set server http://localhost:8080
dot config set actor_id user:alice
dot config set namespace_id ProductTree:/MyProject
dot config set capabilities read,write
```

#### `dot use <namespace>`
Set the active namespace (convenience command).

```bash
dot use ProductTree:/MyProject
```

#### `dot whereami`
Show resolved configuration (including environment overrides).

```bash
dot whereami
```

### Read Commands

#### `dot show <node-id>`
Display a node with its relationships, roles, and links.

```bash
# Show a node
dot show node:my-goal

# Show with depth
dot show node:my-goal --depth 2

# Show at a specific sequence
dot show node:my-goal --asof-seq 42

# Show at a specific time
dot show node:my-goal --asof-time 2024-01-01T00:00:00Z

# Override namespace for this command
dot show node:my-goal --ns ProductTree:/OtherProject

# JSON output
dot show node:my-goal --json
```

#### `dot history <target>`
Get operation history for a node or namespace.

```bash
# Get history for a node
dot history node:my-goal

# Limit results
dot history node:my-goal --limit 10

# Get history for a namespace
dot history ProductTree:/MyProject
```

#### `dot diff <a> <b> <target>`
Show differences between two sequence points.

```bash
# Diff between two sequences
dot diff 10 20 node:my-goal

# Diff from sequence to now
dot diff 10 now node:my-goal

# Diff from now to sequence
dot diff now 20 node:my-goal
```

#### `dot ls <node-id>`
List children of a node (nodes connected via PARENT_OF links).

```bash
# List children
dot ls node:parent-node

# List at specific sequence
dot ls node:parent-node --asof-seq 42
```

### Write Commands

All write commands follow the **plan → apply** workflow:
1. Create a plan from intents
2. Display the plan and policy report
3. Check for policy denies (exit code 2 if found)
4. Prompt for confirmation (unless `--yes`)
5. Apply the plan

#### `dot new node "<title>"`
Create a new node.

```bash
# Create a node
dot new node "My First Goal"

# Create with metadata
dot new node "My Goal" --meta priority=high --meta status=active

# Dry run (plan only, don't apply)
dot new node "My Goal" --dry-run

# Skip confirmation
dot new node "My Goal" --yes

# JSON output
dot new node "My Goal" --json
```

#### `dot role assign <node-id> <role>`
Assign a role to a node.

```bash
# Assign a role
dot role assign node:my-goal Goal

# With flags
dot role assign node:my-goal WorkItem --ns ProductTree:/MyProject --yes
```

#### `dot link <from> <type> <to>`
Create a link between nodes.

```bash
# Create a PARENT_OF link
dot link node:parent PARENT_OF node:child

# Create a RELATED_TO link
dot link node:goal-1 RELATED_TO node:goal-2

# With flags
dot link node:parent PARENT_OF node:child --ns ProductTree:/MyProject --yes
```

#### `dot move <child> --to <parent>`
Move a node to a new parent.

```bash
# Move a node
dot move node:child --to node:new-parent

# With flags
dot move node:child --to node:new-parent --ns ProductTree:/MyProject --yes
```

## Global Flags

All commands support these global flags:

- `--server <url>` - Override server URL
- `--actor <id>` - Override actor ID
- `--ns <namespace>` - Override namespace for this command
- `--cap <capabilities>` - Override capabilities (comma-separated)
- `--json` - Output JSON instead of text
- `-n, --dry-run` - Plan only (mutations, don't apply)
- `-y, --yes` - Skip confirmation (mutations)

## Environment Variables

You can override configuration via environment variables:

```bash
export DOT_SERVER=http://localhost:8080
export DOT_ACTOR=user:alice
export DOT_NAMESPACE=ProductTree:/MyProject
export DOT_CAPABILITIES=read,write

# Now commands use these values
dot status
```

## Example Workflow

```bash
# 1. Configure
dot use ProductTree:/MyProject
dot config set actor_id user:alice

# 2. Check status
dot status

# 3. Create a goal node
dot new node "Launch Product" --yes

# Save the node ID from output (e.g., node:abc123)

# 4. Assign a role
dot role assign node:abc123 Goal --yes

# 5. Create a child work item
dot new node "Design UI" --yes
# Save node ID (e.g., node:def456)

# 6. Link them
dot link node:abc123 PARENT_OF node:def456 --yes

# 7. View the goal with its children
dot show node:abc123 --depth 2

# 8. Check history
dot history node:abc123

# 9. List children
dot ls node:abc123
```

## Exit Codes

- `0` - Success
- `1` - Usage/validation error (client-side)
- `2` - Policy denied
- `3` - Conflict (plan already applied / hash mismatch)
- `4` - Network/server error

## Configuration File

Configuration is stored at `~/.dot/config.json`:

```json
{
  "server": "http://localhost:8080",
  "actor_id": "user:alice",
  "namespace_id": "ProductTree:/MyProject",
  "capabilities": ["read", "write"]
}
```

## Troubleshooting

### "Connection refused" or network errors
- Make sure the kernel server is running: `make dev` in the kernel directory
- Check the server URL: `dot config get server`
- Test connectivity: `dot status`

### "Policy denied" errors
- Check the policy set for your namespace
- Review the plan output to see which rule was violated
- Exit code 2 indicates policy denial

### "Plan hash mismatch" or "already applied"
- A plan can only be applied once
- Create a new plan if you need to make changes
- Exit code 3 indicates conflict

### Configuration not working
- Check config file: `cat ~/.dot/config.json`
- Check environment variables: `env | grep DOT`
- View resolved config: `dot whereami`

## Tips

- Use `--dry-run` to preview changes without applying
- Use `--json` for scripting and automation
- Use `--yes` in scripts to skip confirmation prompts
- Use `dot whereami` to debug configuration issues
- Use `dot status` to verify kernel connectivity

## Help

Get help for any command:

```bash
dot --help
dot show --help
dot new node --help
```
