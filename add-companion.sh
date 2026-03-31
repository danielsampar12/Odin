#!/bin/bash

AICHAT_CONFIG_DIR="${1:-$HOME/.config/aichat}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROLES_DIR="$SCRIPT_DIR/roles"

GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m'

info()   { echo -e "${GREEN}  ✓${NC} $1"; }
prompt() { echo -e "${BLUE}  ?${NC} $1"; }
warn()   { echo -e "${RED}  ⚠${NC} $1"; }

echo ""
echo "  Forge a new companion."
echo ""

prompt "Name (lowercase, e.g. 'freya'): "
read -r COMPANION_NAME
COMPANION_NAME=$(echo "$COMPANION_NAME" | tr '[:upper:]' '[:lower:]' | tr ' ' '-')

if [ -z "$COMPANION_NAME" ]; then
  echo "  No name given. Aborted."
  exit 1
fi

if [[ "$COMPANION_NAME" =~ [^a-z0-9_-] ]]; then
  warn "Name must only contain letters, numbers, hyphens, and underscores."
  exit 1
fi

DEST="$ROLES_DIR/${COMPANION_NAME}.md"

if [ -f "$DEST" ]; then
  warn "A companion named '${COMPANION_NAME}' already exists."
  prompt "Overwrite? (y/N): "
  read -r _ans
  [[ "$_ans" =~ ^[Yy]$ ]] || { echo "  Aborted."; exit 0; }
fi

echo ""
echo "  How do you want to define '${COMPANION_NAME}'?"
echo ""
echo "  1. Path to an existing .md file"
echo "  2. Paste the role text now"
echo ""
prompt "Choose (1 or 2): "
read -r MODE

case "$MODE" in
  1)
    echo ""
    prompt "Path to .md file: "
    read -r FILE_PATH
    FILE_PATH="${FILE_PATH/#\~/$HOME}"
    if [ ! -f "$FILE_PATH" ]; then
      warn "File not found: $FILE_PATH"
      exit 1
    fi
    cp "$FILE_PATH" "$DEST"
    ;;
  2)
    echo ""
    echo "  Paste the role definition below."
    echo "  Describe how this companion thinks, writes, and behaves."
    echo "  Enter a line with just END when you're done."
    echo ""
    ROLE_TEXT=""
    while IFS= read -r line; do
      [ "$line" = "END" ] && break
      ROLE_TEXT="${ROLE_TEXT}${line}"$'\n'
    done
    if [ -z "$(echo "$ROLE_TEXT" | tr -d '[:space:]')" ]; then
      warn "Nothing provided. Aborted."
      exit 1
    fi
    printf '%s' "$ROLE_TEXT" > "$DEST"
    ;;
  *)
    warn "Invalid choice. Aborted."
    exit 1
    ;;
esac

mkdir -p "$AICHAT_CONFIG_DIR/roles"
cp "$DEST" "$AICHAT_CONFIG_DIR/roles/${COMPANION_NAME}.md"

echo ""
info "Companion '${COMPANION_NAME}' is ready"
echo ""
echo "  Summon with:"
echo "    <your-assistant> ${COMPANION_NAME} my-project"
echo ""
