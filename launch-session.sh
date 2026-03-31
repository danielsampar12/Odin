#!/bin/bash
# launch-session.sh <role> <session> <aichat_config_dir>
# Starts an aichat session, auto-detecting RAG if one exists for the session name.

ROLE="$1"
SESSION="$2"
AICHAT_CONFIG_DIR="$3"
RAG_FILE="$AICHAT_CONFIG_DIR/rags/${SESSION}.yaml"

if [ -f "$RAG_FILE" ]; then
  # Get age in days (cross-platform)
  if [ "$(uname -s)" = "Darwin" ]; then
    FILE_EPOCH=$(stat -f %m "$RAG_FILE")
  else
    FILE_EPOCH=$(stat -c %Y "$RAG_FILE")
  fi
  NOW_EPOCH=$(date +%s)
  AGE_DAYS=$(( (NOW_EPOCH - FILE_EPOCH) / 86400 ))

  if [ "$AGE_DAYS" -eq 0 ]; then
    AGE_LABEL="indexed today"
  elif [ "$AGE_DAYS" -eq 1 ]; then
    AGE_LABEL="indexed 1 day ago"
  else
    AGE_LABEL="indexed ${AGE_DAYS} days ago"
  fi

  echo ""
  echo "  RAG found for '${SESSION}' (${AGE_LABEL})"
  echo ""
  echo "  1) Use it"
  echo "  2) Rebuild first"
  echo "  3) Skip"
  echo ""
  printf "  Choose (1/2/3, default: 1): "
  read -r RAG_CHOICE

  case "${RAG_CHOICE:-1}" in
    2)
      echo ""
      echo "  Rebuilding..."
      aichat --rag "$SESSION" --rebuild-rag
      echo ""
      aichat --role "$ROLE" --session "$SESSION" --rag "$SESSION"
      ;;
    3)
      aichat --role "$ROLE" --session "$SESSION"
      ;;
    *)
      aichat --role "$ROLE" --session "$SESSION" --rag "$SESSION"
      ;;
  esac
else
  aichat --role "$ROLE" --session "$SESSION"
fi
