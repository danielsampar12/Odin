#!/bin/bash
# index-project.sh <aichat_config_dir> [rag_name]
# Creates or rebuilds a RAG index for a project directory.

AICHAT_CONFIG_DIR="${1:-$HOME/.config/aichat}"
RAG_NAME="$2"

GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m'

info()   { echo -e "${GREEN}  ✓${NC} $1"; }
prompt() { echo -e "${BLUE}  ?${NC} $1"; }
warn()   { echo -e "${RED}  ⚠${NC} $1"; }

echo ""
echo "  Index a project for RAG."
echo ""

if [ -z "$RAG_NAME" ]; then
  prompt "Index name (default: $(basename "$PWD")): "
  read -r RAG_NAME
  RAG_NAME="${RAG_NAME:-$(basename "$PWD")}"
fi

RAG_FILE="$AICHAT_CONFIG_DIR/rags/${RAG_NAME}.yaml"

# Already exists — just rebuild
if [ -f "$RAG_FILE" ]; then
  echo ""
  echo "  RAG '${RAG_NAME}' already exists. Rebuilding to sync changes..."
  echo ""
  aichat --rag "$RAG_NAME" --rebuild-rag
  echo ""
  info "RAG '${RAG_NAME}' rebuilt"
  exit 0
fi

# New RAG — ask for directory
prompt "Directory to index (default: $PWD): "
read -r PROJECT_DIR
PROJECT_DIR="${PROJECT_DIR:-$PWD}"
PROJECT_DIR="${PROJECT_DIR/#\~/$HOME}"

if [ ! -d "$PROJECT_DIR" ]; then
  warn "Directory not found: $PROJECT_DIR"
  exit 1
fi

# Ask for file extensions
echo ""
echo "  File extensions to include (space-separated)."
echo "  Default: ts tsx js jsx py go rs java md"
echo ""
prompt "Extensions: "
read -r EXTENSIONS
EXTENSIONS="${EXTENSIONS:-ts tsx js jsx py go rs java md}"

# Build glob pattern
PATTERN=$(echo "$EXTENSIONS" | tr ' ' ',')
if [[ "$PATTERN" == *","* ]]; then
  DOC_PATH="${PROJECT_DIR}/**/*.{${PATTERN}}"
else
  DOC_PATH="${PROJECT_DIR}/**/*.${PATTERN}"
fi

# Write RAG YAML
mkdir -p "$AICHAT_CONFIG_DIR/rags"
cat > "$RAG_FILE" << YAML
embedding_model: ollama:nomic-embed-text
chunk_size: 1500
chunk_overlap: 150
reranker_model: null
top_k: 5
batch_size: null
next_file_id: 0
document_paths:
  - "${DOC_PATH}"
files: {}
vectors: {}
YAML

echo ""
echo "  Indexing '${RAG_NAME}' — this may take a while for large projects..."
echo ""
aichat --rag "$RAG_NAME" --rebuild-rag
echo ""
info "RAG '${RAG_NAME}' is ready"
echo ""
echo "  It will be detected automatically when you start a session named '${RAG_NAME}'."
echo ""
