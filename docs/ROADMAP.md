# Odin v2 Roadmap

## Current phase

Foundation and migration safety.

Delivered in the current scaffold:
- preserved the original shell implementation under `legacy/`
- added repo-root compatibility wrappers for the legacy scripts
- introduced a Go module and Cobra-based CLI
- added `odin version`
- added read-only `odin doctor`
- added `odin memory status` and `odin memory doctor`
- added scaffolded `odin setup`
- added scaffolded `odin init`
- added scaffolded `odin start`
- added Ollama model listing, recommendation, and pull scaffolds
- added project-local OpenCode config generation
- added a MemPalace detection and MCP-generation scaffold
- added path helpers, config templates, companion defaults, and plugin detection stubs

## Near-term next steps

### 1. Stronger diagnostics

- detect Ollama API reachability, not just command presence
- surface installed Ollama models
- add clearer OpenCode and MemPalace capability checks
- surface whether generated OpenCode config has explicit MemPalace MCP wiring
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

### 5. Memory integration

- keep MemPalace as Odin's primary intended memory provider
- support explicit OpenCode MCP wiring through project-local generated config
- represent project memory halls in `.odin/config.toml`
- keep Markdown or flat-file exports as fallback, debug, or backup paths only
- shape memory namespaces around:
  - `odin:user/preferences`
  - `odin:user/hardware`
  - `odin:project:<project>/architecture`
  - `odin:project:<project>/decisions`
  - `odin:project:<project>/sessions`
  - `odin:companion:<name>/diary`
  - `odin:companion:<name>/style`

### 6. Companion management

- add `odin companion list`
- add `odin companion inspect`
- add `odin companion install`
- preserve compatibility with the existing role markdown files

### 7. Model management

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
