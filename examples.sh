#!/bin/bash
# Example commands for using the Futurematic Kernel via terminal

BASE_URL="http://localhost:8080"

echo "=== Futurematic Kernel Terminal Examples ==="
echo ""

# Health check
echo "1. Health Check:"
curl -s "$BASE_URL/v1/healthz" | jq . 2>/dev/null || curl -s "$BASE_URL/v1/healthz"
echo -e "\n"

# Create a plan
echo "2. Create a Plan (Create Node):"
PLAN_RESPONSE=$(curl -s -X POST "$BASE_URL/v1/plan" \
  -H "Content-Type: application/json" \
  -d '{
    "actor_id": "user:alice",
    "capabilities": ["read", "write"],
    "namespace_id": "ProductTree:/Example",
    "intents": [
      {
        "kind": "CreateNode",
        "namespace_id": "ProductTree:/Example",
        "payload": {
          "node_id": "node:example-1",
          "title": "Example Node",
          "meta": {}
        }
      }
    ]
  }')

echo "$PLAN_RESPONSE" | jq . 2>/dev/null || echo "$PLAN_RESPONSE"
echo -e "\n"

# Extract plan details (if jq is available)
if command -v jq &> /dev/null; then
  PLAN_ID=$(echo "$PLAN_RESPONSE" | jq -r '.id')
  PLAN_HASH=$(echo "$PLAN_RESPONSE" | jq -r '.hash')
  
  if [ "$PLAN_ID" != "null" ] && [ -n "$PLAN_ID" ]; then
    echo "3. Apply the Plan:"
    APPLY_RESPONSE=$(curl -s -X POST "$BASE_URL/v1/apply" \
      -H "Content-Type: application/json" \
      -d "{
        \"actor_id\": \"user:alice\",
        \"capabilities\": [\"read\", \"write\"],
        \"plan_id\": \"$PLAN_ID\",
        \"plan_hash\": \"$PLAN_HASH\"
      }")
    
    echo "$APPLY_RESPONSE" | jq . 2>/dev/null || echo "$APPLY_RESPONSE"
    echo -e "\n"
    
    echo "4. Expand (Query) the Node:"
    curl -s "$BASE_URL/v1/expand?ids=node:example-1&namespace_id=ProductTree:/Example&depth=1" | jq . 2>/dev/null || curl -s "$BASE_URL/v1/expand?ids=node:example-1&namespace_id=ProductTree:/Example&depth=1"
    echo -e "\n"
  fi
else
  echo "Note: Install 'jq' for better JSON formatting and automatic plan application"
  echo "  macOS: brew install jq"
  echo "  Linux: apt-get install jq  (or equivalent)"
fi
