# Odin v2 Roadmap

## Current phase

Foundation and migration safety.

Delivered in the current scaffold:
- preserved the original shell implementation under `legacy/`
- added repo-root compatibility wrappers for the legacy scripts
- introduced a Go module and Cobra-based CLI
- added `odin version`
- added read-only `odin doctor`
- added scaffolded `odin setup`
- added scaffolded `odin init`
- added scaffolded `odin start`
- added path helpers, config templates, companion defaults, and plugin detection stubs

## Near-term next steps

### 1. Stronger diagnostics

- detect Ollama API reachability, not just command presence
- surface installed Ollama models
- add clearer OpenCode and MemPalace capability checks
- improve shell integration diagnostics for Powerlevel10k

### 2. Setup guidance

- add profile selection for beginner vs developer mode
- add interactive confirmation flow for recommended setup
- add safe install and verify adapters for Ollama, OpenCode, and MemPalace
- register global companions during setup

### 3. Project initialization

- generate richer `.odin/config.toml`
- generate companion-aware `AGENTS.md`
- detect language and framework defaults
- associate projects with MemPalace halls or namespaces
- generate OpenCode project config

### 4. Start and launch flow

- validate global and project config
- ensure Ollama is running
- ensure MemPalace connectivity is available
- refresh Odin status cache
- launch OpenCode with generated config

### 5. Companion management

- add `odin companion list`
- add `odin companion inspect`
- add `odin companion install`
- preserve compatibility with the existing role markdown files

### 6. Model management

- add `odin model recommend`
- add model pull and verify flows
- keep recommendations hardware-aware and local-first

## Explicit non-goals for the MVP

- building a new coding agent loop
- building a new LLM runtime
- implementing deep RAG
- implementing a custom memory database
- implementing a custom prompt engine

## Architecture direction

Odin should remain the glue layer:
- adapters around existing tools
- inspectable generated config
- small, composable packages
- safe and idempotent commands

## Suggested next implementation step

Improve `odin doctor` with:
- Ollama API detection
- model listing
- clearer local runtime health reporting

After that:
- add `odin model recommend`
- add companion registry commands
- add a MemPalace adapter scaffold
- add an OpenCode config generator
