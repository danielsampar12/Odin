#!/bin/bash
set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m'

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
chmod +x "$SCRIPT_DIR/add-companion.sh" "$SCRIPT_DIR/uninstall.sh" "$SCRIPT_DIR/index-project.sh" "$SCRIPT_DIR/launch-session.sh" "$SCRIPT_DIR/setup.sh" 2>/dev/null || true

info()    { echo -e "${GREEN}  ✓${NC} $1"; }
prompt()  { echo -e "${BLUE}  ?${NC} $1"; }
section() { echo -e "\n${YELLOW}==${NC} $1"; }
warn()    { echo -e "${RED}  ⚠${NC} $1"; }

echo ""
echo "  I am Odin — Allfather, god of wisdom and knowledge."
echo "  I hung from Yggdrasil for nine days, wounded by my own spear,"
echo "  no food, no water — until the runes revealed themselves to me."
echo "  I invented writing. You're welcome."
echo ""
echo "  Today, I am running a bash script. For you. On a laptop."
echo "  The sacrifices keep getting harder to justify."
echo ""
echo "  Regardless — I do not do things halfway."
echo "  I will set up your machine, summon your companion, and vanish."
echo "  That is what I do: I connect realms. Even ones running npm."
echo ""
echo "  But first — what shall you call me on this machine?"
echo "  I have over two hundred names across the nine realms."
echo "  Surely you can pick one."
echo ""
prompt "Give me a name (default: odin): "
read -r ASSISTANT_NAME
ASSISTANT_NAME="${ASSISTANT_NAME:-odin}"
ASSISTANT_NAME_LOWER=$(echo "$ASSISTANT_NAME" | tr '[:upper:]' '[:lower:]')
echo ""
info "So be it. I am $ASSISTANT_NAME. The nine realms await."

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

# ── Role selection ────────────────────────────────────────────────────────────
echo ""
echo "  Now — choose your companion. They will ride with you into battle."
echo "  Or, you know, help you fix that bug you've been staring at for three hours."
echo ""
echo "  1. Baldur (pair programmer)"
echo "     Most beloved of all gods. Patient, wise, trusted by everyone in Asgard."
echo "     He will not write your code for you — he will make sure you understand"
echo "     every decision, call out bad patterns, and ask questions before charging in."
echo "     Annoyingly right most of the time. Nobody could hate Baldur."
echo ""
echo "  2. Tyr (architect / reviewer)"
echo "     God of law and justice. He sacrificed his hand to bind the wolf Fenrir —"
echo "     he knew exactly what it would cost and paid it anyway."
echo "     He is not here to implement or teach. He reviews your code from the outside,"
echo "     flags what will hurt you later, and will not let bad decisions through."
echo ""
echo "  3. Thor (implementer)"
echo "     God of thunder. Direct, powerful, and not known for overthinking."
echo "     Point him at a task and get battle-ready code. Fast, no detours."
echo "     Just give him clear instructions. The hammer doesn't ask questions."
echo ""
echo "  4. Loki (chaos agent)"
warn "Not recommended for production. You have been warned."
echo "     Shapeshifter. Trickster. The god who always finds a third option."
echo "     He will solve your problem — just not the way you expected."
echo "     Creative, unconventional, occasionally brilliant, sometimes dangerous."
echo "     Do not give him vague tasks. He will interpret them creatively."
echo ""
prompt "Choose your companion (1, 2, 3 or 4, default: 1): "
read -r ROLE_CHOICE

case "${ROLE_CHOICE:-1}" in
  2) ROLE="tyr";    ROLE_DISPLAY="Tyr" ;;
  3) ROLE="thor";   ROLE_DISPLAY="Thor" ;;
  4) ROLE="loki";   ROLE_DISPLAY="Loki"
     echo ""
     warn "You chose Loki. Bold. Review everything he gives you."
     ;;
  *) ROLE="baldur"; ROLE_DISPLAY="Baldur" ;;
esac

info "$ROLE_DISPLAY will ride with you"

# ── Sessions dir ──────────────────────────────────────────────────────────────
SESSIONS_DIR="$HOME/ai/sessions"
mkdir -p "$SESSIONS_DIR"

# ── Ollama ────────────────────────────────────────────────────────────────────
section "Ollama"
if command -v ollama &>/dev/null; then
  info "Ollama already installed"
else
  echo "  Installing Ollama..."
  if [ "$PLATFORM" = "linux" ]; then
    curl -fsSL https://ollama.com/install.sh | sh
  elif [ "$PLATFORM" = "macos" ]; then
    if command -v brew &>/dev/null; then
      brew install ollama
    else
      echo "  Homebrew not found. Install it first: https://brew.sh"
      exit 1
    fi
  fi
  info "Ollama installed"
fi

# Configure Ollama
if [ "$PLATFORM" = "linux" ]; then
  sudo mkdir -p /etc/systemd/system/ollama.service.d
  sudo tee /etc/systemd/system/ollama.service.d/override.conf > /dev/null << 'EOF'
[Service]
Environment="OLLAMA_FLASH_ATTENTION=1"
Environment="OLLAMA_KV_CACHE_TYPE=q8_0"
Environment="OLLAMA_KEEP_ALIVE=15m"
Environment="OLLAMA_CONTEXT_LENGTH=32768"
Environment="OLLAMA_NO_CLOUD=1"
EOF
  sudo systemctl daemon-reload
  sudo systemctl disable ollama 2>/dev/null || true
  sudo systemctl start ollama
  info "Ollama configured (systemd, disabled on boot)"

elif [ "$PLATFORM" = "macos" ]; then
  if ! grep -q "OLLAMA_NO_CLOUD" "$SHELL_RC" 2>/dev/null; then
    cat >> "$SHELL_RC" << 'EOF'

# Ollama
export OLLAMA_FLASH_ATTENTION=1
export OLLAMA_KV_CACHE_TYPE=q8_0
export OLLAMA_KEEP_ALIVE=15m
export OLLAMA_CONTEXT_LENGTH=32768
export OLLAMA_NO_CLOUD=1
EOF
  fi
  if ! pgrep -x ollama &>/dev/null; then
    ollama serve > /dev/null 2>&1 &
    sleep 2
  fi
  info "Ollama configured"
fi

# ── Hardware detection + model recommendation ─────────────────────────────────
section "Hardware"
RECOMMENDED_MODEL=""
SAFE_MODEL=""

if [ "$PLATFORM" = "linux" ]; then
  if command -v nvidia-smi &>/dev/null; then
    VRAM=$(nvidia-smi --query-gpu=memory.total --format=csv,noheader,nounits 2>/dev/null | head -1)
    VRAM_GB=$((VRAM / 1024))
    echo "  NVIDIA GPU detected — ${VRAM_GB}GB VRAM"
    # Compare in MiB (raw value) to avoid integer division rounding issues.
    # Thresholds sit in the gap between card tiers — a 12GB card always reports
    # >11,000 MiB; a 24GB card always reports >23,000 MiB; an 8GB card never
    # exceeds 8,192 MiB. This is driver-overhead-agnostic.
    if   [ "$VRAM" -ge 23000 ]; then RECOMMENDED_MODEL="qwen3-coder:30b";                   SAFE_MODEL="qwen2.5-coder:32b"
    elif [ "$VRAM" -ge 10240 ]; then RECOMMENDED_MODEL="qwen2.5-coder:14b-instruct-q5_K_M"; SAFE_MODEL="qwen2.5-coder:14b-instruct-q4_K_M"
    elif [ "$VRAM" -ge 6144  ]; then RECOMMENDED_MODEL="qwen2.5-coder:7b";                  SAFE_MODEL="qwen2.5-coder:7b"
    else                              RECOMMENDED_MODEL="qwen2.5-coder:3b";                  SAFE_MODEL="qwen2.5-coder:3b"
    fi
  else
    RAM=$(free -g | awk '/^Mem:/{print $2}')
    echo "  No NVIDIA GPU — RAM: ${RAM}GB (CPU inference, will be slow)"
    if   [ "$RAM" -ge 16 ]; then RECOMMENDED_MODEL="qwen2.5-coder:7b"; SAFE_MODEL="qwen2.5-coder:3b"
    else                         RECOMMENDED_MODEL="qwen2.5-coder:3b"; SAFE_MODEL="qwen2.5-coder:3b"
    fi
  fi

elif [ "$PLATFORM" = "macos" ]; then
  RAM=$(( $(sysctl -n hw.memsize) / 1024 / 1024 / 1024 ))
  if [[ "$(uname -m)" == "arm64" ]]; then
    echo "  Apple Silicon — unified memory: ${RAM}GB"
    if   [ "$RAM" -ge 48 ]; then RECOMMENDED_MODEL="qwen3-coder:30b";                   SAFE_MODEL="qwen2.5-coder:32b"
    elif [ "$RAM" -ge 32 ]; then RECOMMENDED_MODEL="qwen3-coder:30b";                   SAFE_MODEL="qwen2.5-coder:14b-instruct-q5_K_M"
    elif [ "$RAM" -ge 16 ]; then RECOMMENDED_MODEL="qwen2.5-coder:14b-instruct-q5_K_M"; SAFE_MODEL="qwen2.5-coder:14b-instruct-q4_K_M"
    else                         RECOMMENDED_MODEL="qwen2.5-coder:7b";                  SAFE_MODEL="qwen2.5-coder:3b"
    fi
  else
    echo "  Intel Mac — RAM: ${RAM}GB (limited GPU acceleration)"
    RECOMMENDED_MODEL="qwen2.5-coder:7b"
    SAFE_MODEL="qwen2.5-coder:3b"
  fi
fi

echo ""
echo "  Recommended  : $RECOMMENDED_MODEL"
echo "  Safe fallback: $SAFE_MODEL"
echo ""
prompt "Use recommended model? (Y/n): "
read -r USE_RECOMMENDED
if [[ "$USE_RECOMMENDED" =~ ^[Nn]$ ]]; then
  prompt "Use safe fallback? (Y/n): "
  read -r USE_SAFE
  if [[ "$USE_SAFE" =~ ^[Nn]$ ]]; then
    prompt "Enter model name manually: "
    read -r MODEL_NAME
  else
    MODEL_NAME="$SAFE_MODEL"
  fi
else
  MODEL_NAME="$RECOMMENDED_MODEL"
fi

# ── Pull model ────────────────────────────────────────────────────────────────
section "Model"
echo "  Pulling $MODEL_NAME — this will take a while..."
ollama pull "$MODEL_NAME"
info "Model pulled"

# Build coder model from Modelfile + chosen role as system prompt
SYSTEM_PROMPT=$(cat "$SCRIPT_DIR/roles/$ROLE.md")
cat > /tmp/coder.Modelfile << EOF
FROM $MODEL_NAME
PARAMETER temperature 0.3
PARAMETER num_ctx 32768
PARAMETER num_predict 4096
SYSTEM """
$SYSTEM_PROMPT
"""
EOF

ollama create coder -f /tmp/coder.Modelfile
rm /tmp/coder.Modelfile
info "Model 'coder' created with $ROLE_DISPLAY's personality"

echo "  Pulling embedding model for RAG (nomic-embed-text)..."
ollama pull nomic-embed-text
info "Embedding model ready"

# ── aichat ────────────────────────────────────────────────────────────────────
section "aichat"
if command -v aichat &>/dev/null; then
  info "aichat already installed"
else
  echo "  Installing aichat..."
  if command -v brew &>/dev/null; then
    brew install aichat
  else
    curl -fsSL https://github.com/sigoden/aichat/releases/latest/download/aichat-x86_64-unknown-linux-musl.tar.gz | tar xz
    sudo mv aichat /usr/local/bin/
  fi
  info "aichat installed"
fi

# aichat config dir differs by platform
if [ "$PLATFORM" = "macos" ]; then
  AICHAT_CONFIG_DIR="$HOME/Library/Application Support/aichat"
else
  AICHAT_CONFIG_DIR="$HOME/.config/aichat"
fi

mkdir -p "$AICHAT_CONFIG_DIR/roles"

# Copy all roles so the user can switch later
cp "$SCRIPT_DIR/roles/baldur.md" "$AICHAT_CONFIG_DIR/roles/baldur.md"
cp "$SCRIPT_DIR/roles/tyr.md"    "$AICHAT_CONFIG_DIR/roles/tyr.md"
cp "$SCRIPT_DIR/roles/thor.md"   "$AICHAT_CONFIG_DIR/roles/thor.md"
cp "$SCRIPT_DIR/roles/loki.md"   "$AICHAT_CONFIG_DIR/roles/loki.md"

cat > "$AICHAT_CONFIG_DIR/config.yaml" << EOF
model: ollama:coder
stream: true
save: true
save_session: true
sessions_dir: $SESSIONS_DIR
rag_embedding_model: ollama:nomic-embed-text

clients:
  - type: openai-compatible
    name: ollama
    api_base: http://localhost:11434/v1
    models:
      - name: coder
        max_input_tokens: 32768
EOF

info "aichat configured"

# ── Shell function ────────────────────────────────────────────────────────────
section "Shell"

if grep -q "^${ASSISTANT_NAME_LOWER}()" "$SHELL_RC" 2>/dev/null; then
  echo "  Found existing function, replacing..."
  if [ "$PLATFORM" = "macos" ]; then
    sed -i '' "/^${ASSISTANT_NAME_LOWER}()/,/^}/d" "$SHELL_RC"
  else
    sed -i "/^${ASSISTANT_NAME_LOWER}()/,/^}/d" "$SHELL_RC"
  fi
fi

if [ "$PLATFORM" = "linux" ]; then
  START_CMD="sudo systemctl start ollama && echo '${ASSISTANT_NAME} awakens'"
  STOP_CMD="sudo systemctl stop ollama && echo '${ASSISTANT_NAME} returns to Asgard'"
elif [ "$PLATFORM" = "macos" ]; then
  START_CMD="ollama serve > /dev/null 2>&1 & echo '${ASSISTANT_NAME} awakens'"
  STOP_CMD="pkill ollama && echo '${ASSISTANT_NAME} returns to Asgard'"
fi

cat >> "$SHELL_RC" << EOF

# ${ASSISTANT_NAME}
${ASSISTANT_NAME_LOWER}() {
  case "\$1" in
    start)  ${START_CMD} ;;
    stop)   ${STOP_CMD} ;;
    new)    aichat --role $ROLE ;;
    list)   aichat --list-sessions ;;
    add)    bash "$SCRIPT_DIR/add-companion.sh" "$AICHAT_CONFIG_DIR" ;;
    index)  bash "$SCRIPT_DIR/index-project.sh" "$AICHAT_CONFIG_DIR" "\$2" ;;
    remove)
      if [ -z "\$2" ]; then
        echo "  Usage: ${ASSISTANT_NAME_LOWER} remove <session>"
      elif [ -f "${AICHAT_CONFIG_DIR}/sessions/\${2}.yaml" ]; then
        printf "  Remove session '\$2'? (y/N): "
        read -r _ans
        if [[ "\$_ans" =~ ^[Yy]\$ ]]; then
          rm "${AICHAT_CONFIG_DIR}/sessions/\${2}.yaml"
          echo "  Session '\$2' removed."
        fi
      else
        echo "  Session '\$2' not found. Run '${ASSISTANT_NAME_LOWER} list' to see available sessions."
      fi
      ;;
    baldur) bash "$SCRIPT_DIR/launch-session.sh" baldur "\${2:-default}" "$AICHAT_CONFIG_DIR" ;;
    tyr)    bash "$SCRIPT_DIR/launch-session.sh" tyr    "\${2:-default}" "$AICHAT_CONFIG_DIR" ;;
    thor)   bash "$SCRIPT_DIR/launch-session.sh" thor   "\${2:-default}" "$AICHAT_CONFIG_DIR" ;;
    loki)   bash "$SCRIPT_DIR/launch-session.sh" loki   "\${2:-default}" "$AICHAT_CONFIG_DIR" ;;
    *)
      if [ -f "${AICHAT_CONFIG_DIR}/roles/\${1}.md" ]; then
        bash "$SCRIPT_DIR/launch-session.sh" "\$1" "\${2:-default}" "$AICHAT_CONFIG_DIR"
      else
        bash "$SCRIPT_DIR/launch-session.sh" $ROLE "\${1:-default}" "$AICHAT_CONFIG_DIR"
      fi
      ;;
  esac
}
EOF

info "Shell function '${ASSISTANT_NAME_LOWER}' added to $SHELL_RC"

# ── Done ──────────────────────────────────────────────────────────────────────
echo ""
echo "  My work here is done."
echo "  $ROLE_DISPLAY is ready. Go build something worthy of the nine realms."
echo "  (Or at least something that passes the tests. One step at a time.)"
echo ""
echo "  Reload your shell first:"
echo "    source $SHELL_RC"
echo ""
echo "  Then:"
echo "    ${ASSISTANT_NAME_LOWER} start          # awaken"
echo "    ${ASSISTANT_NAME_LOWER} my-project     # start or resume a session"
echo "    ${ASSISTANT_NAME_LOWER} baldur my-proj # summon Baldur"
echo "    ${ASSISTANT_NAME_LOWER} tyr my-proj    # summon Tyr"
echo "    ${ASSISTANT_NAME_LOWER} thor my-proj   # summon Thor"
echo "    ${ASSISTANT_NAME_LOWER} loki my-proj   # summon Loki (you were warned)"
echo "    ${ASSISTANT_NAME_LOWER} list             # see all sessions"
echo "    ${ASSISTANT_NAME_LOWER} index my-proj   # index a project for RAG"
echo "    ${ASSISTANT_NAME_LOWER} remove my-proj  # remove a session"
echo "    ${ASSISTANT_NAME_LOWER} add             # forge a new companion"
echo "    ${ASSISTANT_NAME_LOWER} stop            # return to Asgard"
echo ""
