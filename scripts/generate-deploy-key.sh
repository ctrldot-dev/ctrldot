#!/usr/bin/env bash
# Generate an SSH key for GitHub (deploy key or account SSH key).
# Keys are written to keys/ (gitignored). Add the PUBLIC key to GitHub.
set -e
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
KEY_DIR="$REPO_ROOT/keys"
KEY_FILE="$KEY_DIR/ctrldot_github_deploy"

mkdir -p "$KEY_DIR"
chmod 700 "$KEY_DIR"

if [ -f "$KEY_FILE" ]; then
  echo "Key already exists: $KEY_FILE"
  echo "Public key (add this to GitHub):"
  echo "---"
  cat "$KEY_FILE.pub"
  echo "---"
  exit 0
fi

ssh-keygen -t ed25519 -f "$KEY_FILE" -N "" -C "ctrldot-deploy"

echo ""
echo "Created: $KEY_FILE (private) and $KEY_FILE.pub (public)"
echo ""
echo "Add the PUBLIC key to GitHub:"
echo "  • Repo deploy key: Settings → Deploy keys → Add deploy key"
echo "  • Or account SSH key: Settings → SSH and GPG keys → New SSH key"
echo ""
echo "Public key (copy everything below):"
echo "---"
cat "$KEY_FILE.pub"
echo "---"
echo ""
echo "Keep the private key ($KEY_FILE) secret. Do not commit it. (keys/ is in .gitignore.)"
