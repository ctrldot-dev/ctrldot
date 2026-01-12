# Dot CLI Examples

## Basic Node Operations

### Create and View a Node

\`\`\`bash
# Create a node
dot new node "My Goal" --yes

# Output shows the node ID, e.g., node:abc123
# View it
dot show node:abc123
\`\`\`

### Create Node with Metadata

\`\`\`bash
dot new node "High Priority Task" \
  --meta priority=high \
  --meta status=active \
  --meta owner=alice \
  --yes
\`\`\`

## Role Management

### Assign Roles

\`\`\`bash
# Assign Goal role
dot role assign node:abc123 Goal --yes

# Assign WorkItem role
dot role assign node:def456 WorkItem --yes
\`\`\`

## Link Management

### Create Parent-Child Relationships

\`\`\`bash
# Create parent
dot new node "Project" --yes
# Save ID: PARENT_ID=node:parent123

# Create child
dot new node "Task" --yes
# Save ID: CHILD_ID=node:child456

# Link them
dot link $PARENT_ID PARENT_OF $CHILD_ID --yes
\`\`\`

### Create Related Links

\`\`\`bash
# Link two goals as related
dot link node:goal1 RELATED_TO node:goal2 --yes
\`\`\`

## Moving Nodes

### Move a Node to New Parent

\`\`\`bash
# Move child from one parent to another
dot move node:child --to node:new-parent --yes
\`\`\`

## Querying

### Show Node with Depth

\`\`\`bash
# Show node with immediate relationships
dot show node:abc123 --depth 1

# Show node with 2 levels deep
dot show node:abc123 --depth 2
\`\`\`

### Time Travel Queries

\`\`\`bash
# Show node at specific sequence
dot show node:abc123 --asof-seq 10

# Show node at specific time
dot show node:abc123 --asof-time 2024-01-01T00:00:00Z
\`\`\`

### History

\`\`\`bash
# Get full history
dot history node:abc123

# Get last 10 operations
dot history node:abc123 --limit 10

# Get namespace history
dot history ProductTree:/MyProject
\`\`\`

### Differences

\`\`\`bash
# Compare two sequences
dot diff 1 10 node:abc123

# Compare from sequence to now
dot diff 1 now node:abc123

# Compare from now to sequence
dot diff now 10 node:abc123
\`\`\`

## Workflows

### Create a Project Hierarchy

\`\`\`bash
# 1. Create project
PROJECT_ID=$(dot new node "My Project" --json --yes | jq -r '.operation.changes[0].payload.node_id')
dot role assign $PROJECT_ID Project --yes

# 2. Create goals
GOAL1_ID=$(dot new node "Goal 1" --json --yes | jq -r '.operation.changes[0].payload.node_id')
GOAL2_ID=$(dot new node "Goal 2" --json --yes | jq -r '.operation.changes[0].payload.node_id')

dot role assign $GOAL1_ID Goal --yes
dot role assign $GOAL2_ID Goal --yes

# 3. Link goals to project
dot link $PROJECT_ID PARENT_OF $GOAL1_ID --yes
dot link $PROJECT_ID PARENT_OF $GOAL2_ID --yes

# 4. View the hierarchy
dot show $PROJECT_ID --depth 2
\`\`\`

### Script with Error Handling

\`\`\`bash
#!/bin/bash
set -e

# Create node
OUTPUT=$(dot new node "Test Node" --json --yes)
NODE_ID=$(echo $OUTPUT | jq -r '.operation.changes[0].payload.node_id')

if [ "$NODE_ID" = "null" ] || [ -z "$NODE_ID" ]; then
  echo "Failed to create node"
  exit 1
fi

echo "Created node: $NODE_ID"

# Assign role
if ! dot role assign $NODE_ID Goal --yes; then
  echo "Failed to assign role"
  exit 1
fi

echo "Success!"
\`\`\`

## Advanced Usage

### Using Different Namespaces

\`\`\`bash
# Work in one namespace
dot use ProductTree:/ProjectA
dot new node "Node A" --yes

# Work in another namespace (override for one command)
dot new node "Node B" --ns ProductTree:/ProjectB --yes

# Check where you are
dot whereami
\`\`\`

### Dry Run for Safety

\`\`\`bash
# Preview what would happen
dot new node "Test" --dry-run

# Review the plan, then apply manually
dot new node "Test" --yes
\`\`\`

### JSON for Automation

\`\`\`bash
# Get node ID from JSON output
NODE_ID=$(dot new node "Test" --json --yes | jq -r '.operation.changes[0].payload.node_id')

# Use in next command
dot show $NODE_ID --json | jq '.result.nodes[0]'
\`\`\`
