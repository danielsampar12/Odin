# Legacy Odin Shell Flow

This directory preserves Odin's original shell-based implementation.

Contents:
- `setup.sh`: interactive machine bootstrap for Ollama + aichat
- `launch-session.sh`: session launcher with optional RAG detection
- `index-project.sh`: project indexing for the legacy RAG flow
- `add-companion.sh`: custom companion installer for the legacy role system
- `uninstall.sh`: legacy teardown script
- `coder.Modelfile.example`: shell-era model template example

Notes:
- Repo-root wrapper scripts remain in place so older entrypoints like `./setup.sh` still forward into this directory.
- Legacy scripts continue to use the preserved role files from the repo-root `roles/` directory.
- Odin v2 development lives in the Go CLI under `cmd/odin` and `internal/`.
