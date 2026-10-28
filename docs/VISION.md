# Odin v2 Vision

This document describes direction. [ROADMAP.md](ROADMAP.md) tracks what exists today and the decisions taken so far.

## What Odin v2 is

Odin v2 is a local-first AI stack manager.

Its job is to install, configure, validate, and launch a coherent local AI workstation by orchestrating existing tools:
- Odin CLI: control plane, manager, launcher
- Ollama: local model runtime
- OpenCode: coding assistant
- MemPalace: optional persistent local memory (deferred, off by default)
- Powerlevel10k first, Starship later: shell prompt and status integration (deferred)
- Companions: Odin-managed prompts, defaults, and identity

The product goal is simple:

> Odin transforms a set of local AI tools into a coherent local-first AI workstation.

## What Odin v2 is not

Odin v2 should not reinvent the surrounding ecosystem.

It is not:
- a new coding agent
- a new LLM runtime
- a new vector database
- a replacement for OpenCode
- a replacement for Ollama
- a replacement for MemPalace
- a replacement for Starship or Powerlevel10k

It should not implement its own agent loop, diff engine, LSP, memory database, prompt renderer, or full RAG system in the first version.

## Product principles

- Keep local-first and privacy-first defaults.
- Preserve beginner-friendly setup.
- Keep power-user escape hatches.
- Prefer adapters over hard coupling.
- Keep generated configs inspectable.
- Preserve the original repository work.
- Keep the first Go implementation small, safe, and reviewable.

## Core commands

### `odin setup`

Machine-level setup.

This command is expected to:
- detect OS, shell, GPU, RAM, and local tooling
- verify Ollama and OpenCode, and show their install commands instead of installing them
- optionally configure shell integration
- recommend and pull a local model
- create `~/.odin/config.toml`
- register global companions

It should not depend on being inside a project.

### `odin init`

Project-level initialization.

This command is expected to:
- detect the project and its stack
- create `.odin/config.toml`
- create `.odin/rules.md`
- create `AGENTS.md` when the repository has none (planned: only when asked)
- choose a project default companion
- record a MemPalace hall when memory is enabled
- prepare the project for `odin opencode generate`

### `odin start`

Project startup and launch.

This command is expected to:
- load global and project config
- validate the stack
- ensure Ollama is running
- when memory is enabled, check MemPalace or MCP connectivity
- run the model with a context large enough for OpenCode's prompt
- check that the generated OpenCode config is current
- refresh lightweight Odin status state
- launch OpenCode

### `odin doctor`

Read-only diagnostics for the current machine and project.

## Companions

Companions remain a core part of Odin's identity.

Current and planned companion concepts:
- Baldur: pragmatic pair programmer
- Tyr: architect and reviewer
- Thor: fast implementer
- Loki: chaos agent (unconventional approaches)
- Freya: beginner-friendly teacher
- Hephaestus: infra and local setup expert

In v2, Odin should eventually translate companion choices into generated OpenCode or project guidance files without deleting or discarding the existing role files.

## Memory direction

MemPalace is the preferred memory provider when memory is enabled. Memory integration is deferred: it is off by default and opt-in (`provider = "mempalace"` under `[memory]`), until MemPalace itself has been validated.

The product metaphor fits naturally:
- Odin
- Valhalla
- halls
- companions
- persistent memory

Markdown or plain local files can exist as fallback, export, debug, or backup paths, but they should not be presented as the primary long-term product direction.

Future conceptual namespaces include:
- `odin:user/preferences`
- `odin:user/hardware`
- `odin:project:<project>/architecture`
- `odin:project:<project>/decisions`
- `odin:project:<project>/sessions`
- `odin:companion:<name>/diary`
- `odin:companion:<name>/style`

## Shell direction

Shell integration is deferred. When it comes, it should be adapter-based.

Priority:
- Powerlevel10k first
- Starship later

Future prompt rendering such as `odin status --prompt` should be fast and cache-backed rather than running heavy checks on every prompt render.

## Repository strategy

The original shell implementation remains preserved under `legacy/`.

Odin v2 development now lives in the Go CLI layout:
- `cmd/odin`
- `internal/cli`
- `internal/config`
- `internal/doctor`
- `internal/system`
- `internal/plugins`
- `internal/companions`
