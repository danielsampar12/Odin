#!/bin/bash
set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m'

info()    { echo -e "${GREEN}  ✓${NC} $1"; }
prompt()  { echo -e "${BLUE}  ?${NC} $1"; }
section() { echo -e "\n${YELLOW}==${NC} $1"; }
warn()    { echo -e "${RED}  ⚠${NC} $1"; }
removed() { echo -e "${RED}  ✗${NC} $1"; }
ask_no()  { prompt "$1 (y/N): "; read -r _ans; [[ "$_ans" =~ ^[Yy]$ ]]; }

# ── OS detection ──────────────────────────────────────────────────────────────
OS="$(uname -s)"
case "$OS" in
  Darwin) PLATFORM="macos" ;;
  Linux)  PLATFORM="linux" ;;
  *)      echo "Unsupported OS: $OS"; exit 1 ;;
esac

# ── Shell detection ───────────────────────────────────────────────────────────
if [ -f "$HOME/.zshrc" ] && ([ "$SHELL" = "$(which zsh 2>/dev/null)" ] || [ -n "$ZSH_VERSION" ]); then
  SHELL_RC="$HOME/.zshrc"
else
  SHELL_RC="$HOME/.bashrc"
fi

# ── aichat config dir ─────────────────────────────────────────────────────────
if [ "$PLATFORM" = "macos" ]; then
  AICHAT_CONFIG_DIR="$HOME/Library/Application Support/aichat"
else
  AICHAT_CONFIG_DIR="$HOME/.config/aichat"
fi

echo ""
echo "  What do you want to remove?"
echo ""

# ── Assistant name ────────────────────────────────────────────────────────────
prompt "What did you name your assistant? (default: odin): "
read -r ASSISTANT_NAME
ASSISTANT_NAME="${ASSISTANT_NAME:-odin}"
ASSISTANT_NAME_LOWER=$(echo "$ASSISTANT_NAME" | tr '[:upper:]' '[:lower:]')

echo ""
REMOVE_OLLAMA=false
REMOVE_MODELS=false
REMOVE_AICHAT=false
REMOVE_SESSIONS=false

prompt "Ollama (service + binary)? (y/n): ";                                      read -r _a; [[ "$_a" =~ ^[Yy]$ ]] && REMOVE_OLLAMA=true
ask_no  "Downloaded models (~10–20GB, cannot be undone)"                          && REMOVE_MODELS=true
prompt "aichat (binary + config)? (y/n): ";                                       read -r _a; [[ "$_a" =~ ^[Yy]$ ]] && REMOVE_AICHAT=true
ask_no  "Chat sessions (your history, cannot be undone)"                          && REMOVE_SESSIONS=true

echo ""
warn "The shell function for '${ASSISTANT_NAME_LOWER}' will always be removed."
echo ""
prompt "Continue? (y/N): "
read -r CONFIRM
if [[ ! "$CONFIRM" =~ ^[Yy]$ ]]; then
  echo "  Aborted. Nothing was changed."
  exit 0
fi

# ── Ollama ────────────────────────────────────────────────────────────────────
if $REMOVE_OLLAMA; then
  section "Ollama"

  # Remove coder model first if Ollama is still around
  if command -v ollama &>/dev/null; then
    if ollama list 2>/dev/null | grep -q "^coder"; then
      ollama rm coder 2>/dev/null && removed "Model 'coder' removed" || true
    fi
  fi

  if [ "$PLATFORM" = "linux" ]; then
    sudo systemctl stop ollama 2>/dev/null || true
    sudo systemctl disable ollama 2>/dev/null || true

    [ -d /etc/systemd/system/ollama.service.d ] && \
      sudo rm -rf /etc/systemd/system/ollama.service.d && \
      removed "Systemd override removed"

    [ -f /etc/systemd/system/ollama.service ] && \
      sudo rm /etc/systemd/system/ollama.service && \
      removed "Systemd service removed"

    [ -f /usr/local/bin/ollama ] && \
      sudo rm /usr/local/bin/ollama && \
      removed "Ollama binary removed"

    sudo systemctl daemon-reload

  elif [ "$PLATFORM" = "macos" ]; then
    pkill ollama 2>/dev/null || true

    if command -v brew &>/dev/null && brew list ollama &>/dev/null 2>&1; then
      brew uninstall ollama && removed "Ollama uninstalled via Homebrew"
    elif [ -f /usr/local/bin/ollama ]; then
      sudo rm /usr/local/bin/ollama && removed "Ollama binary removed"
    fi

    if grep -q "OLLAMA_NO_CLOUD" "$SHELL_RC" 2>/dev/null; then
      sed -i '' '/# Ollama/,/OLLAMA_NO_CLOUD/d' "$SHELL_RC"
      removed "Ollama env vars removed from $SHELL_RC"
    fi
  fi

  info "Ollama removed"
fi

# ── Models ────────────────────────────────────────────────────────────────────
if $REMOVE_MODELS; then
  section "Models"
  rm -rf "$HOME/.ollama"
  removed "~/.ollama removed"
fi

# ── aichat ────────────────────────────────────────────────────────────────────
if $REMOVE_AICHAT; then
  section "aichat"

  [ -d "$AICHAT_CONFIG_DIR" ] && \
    rm -rf "$AICHAT_CONFIG_DIR" && \
    removed "aichat config removed"

  if command -v brew &>/dev/null && brew list aichat &>/dev/null 2>&1; then
    brew uninstall aichat && removed "aichat uninstalled via Homebrew"
  elif [ -f /usr/local/bin/aichat ]; then
    sudo rm /usr/local/bin/aichat && removed "aichat binary removed"
  fi

  info "aichat removed"
fi

# ── Sessions ──────────────────────────────────────────────────────────────────
if $REMOVE_SESSIONS; then
  section "Sessions"
  [ -d "$HOME/ai/sessions" ] && rm -rf "$HOME/ai/sessions" && removed "Sessions removed"
fi

# ── Shell function (always removed) ──────────────────────────────────────────
section "Shell"
if grep -q "^${ASSISTANT_NAME_LOWER}()" "$SHELL_RC" 2>/dev/null; then
  if [ "$PLATFORM" = "macos" ]; then
    sed -i '' "/^# ${ASSISTANT_NAME}/,/^}/d" "$SHELL_RC"
  else
    sed -i "/^# ${ASSISTANT_NAME}/,/^}/d" "$SHELL_RC"
  fi
  removed "'${ASSISTANT_NAME_LOWER}' function removed from $SHELL_RC"
else
  info "No shell function found for '${ASSISTANT_NAME_LOWER}'"
fi

# ── Done ──────────────────────────────────────────────────────────────────────
echo ""
echo "  All done. The nine realms are clean."
echo ""
echo "  Reload your shell:"
echo "    source $SHELL_RC"
echo ""
