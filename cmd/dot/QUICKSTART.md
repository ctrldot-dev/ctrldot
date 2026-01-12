# Dot CLI Quick Start

## 1. Build the CLI

```bash
# From the project root
go build -o bin/dot ./cmd/dot

# Make it executable and add to PATH (optional)
chmod +x bin/dot
export PATH=$PATH:$(pwd)/bin
```

## 2. Start the Kernel Server

In a separate terminal, start the kernel:

```bash
# From the project root
make dev
```

Wait for it to start (you'll see "Kernel server started successfully").

## 3. Configure Dot CLI

```bash
# Set your namespace
dot use ProductTree:/MyProject

# Set your actor ID
dot config set actor_id user:alice

# Verify configuration
dot whereami
```

## 4. Test Connection

```bash
# Check if kernel is reachable
dot status
```

Expected output:
```
server=http://localhost:8080 actor=user:alice namespace=ProductTree:/MyProject ok=true
```

## 5. Create Your First Node

```bash
# Create a goal node
dot new node "Launch Product" --yes
```

You'll see output like:
```
PLAN plan:... hash=sha256:... class=1
CHANGES:
  + CreateNode node:abc123 title="Launch Product"
APPLIED op:... seq=1 occurred_at=...
CHANGES:
  + CreateNode node:abc123 title="Launch Product"
```

**Save the node ID** (e.g., `node:abc123`) from the output.

## 6. Assign a Role

```bash
# Assign Goal role to the node
dot role assign node:abc123 Goal --yes
```

## 7. View the Node

```bash
# Show the node with its relationships
dot show node:abc123
```

## 8. Create a Child Node

```bash
# Create a work item
dot new node "Design UI" --yes
# Save the node ID (e.g., node:def456)

# Link it as a child
dot link node:abc123 PARENT_OF node:def456 --yes
```

## 9. Explore

```bash
# Show the parent with children
dot show node:abc123 --depth 2

# List children
dot ls node:abc123

# View history
dot history node:abc123

# View differences
dot diff 1 now node:abc123
```

## Common Patterns

### Create a Hierarchy

```bash
# Create parent
dot new node "Project" --yes
# Save ID as PARENT_ID

# Create children
dot new node "Task 1" --yes
dot new node "Task 2" --yes
# Save IDs as CHILD1_ID, CHILD2_ID

# Link them
dot link $PARENT_ID PARENT_OF $CHILD1_ID --yes
dot link $PARENT_ID PARENT_OF $CHILD2_ID --yes
```

### Dry Run (Preview Changes)

```bash
# See what would happen without applying
dot new node "Test Node" --dry-run
```

### JSON Output for Scripting

```bash
# Get JSON output
dot show node:abc123 --json | jq .

# Use in scripts
NODE_ID=$(dot new node "Test" --json --yes | jq -r '.operation.changes[0].payload.node_id')
echo "Created node: $NODE_ID"
```

### Override Configuration Per Command

```bash
# Use different namespace for one command
dot show node:abc123 --ns ProductTree:/OtherProject

# Use different server
dot status --server http://other-server:8080
```

## Troubleshooting

**"Connection refused"**
- Make sure kernel is running: `make dev`
- Check server URL: `dot config get server`

**"Policy denied"**
- Check the plan output to see which rule failed
- Review namespace policy set

**"Plan hash mismatch"**
- Plans can only be applied once
- Create a new plan if needed

## Next Steps

- Read the full [README.md](README.md) for all commands
- Explore `dot --help` for command details
- Check [TESTING.md](TESTING.md) for test information
