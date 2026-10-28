# Odin v2 Roadmap

Updated 2026-10-28.

**Next milestone: Odin works for someone who isn't its creator.** Three people on fresh machines (at least one macOS and one Linux) reach a successful, tool-using OpenCode edit through `odin start --exec`, using only the README, within 45 minutes and without help.

## Done

- Go CLI with `doctor`, `setup`, `init`, `start`, `model list/recommend/pull`, `opencode status/generate` and `memory status/doctor`.
- **Stage 0, core loop proven** on an RTX 5070 (12 GB): `qwen3.5:9b-q4_K_M` at a 16k–32k context completed the OpenCode test task in 2 of 3 runs. qwen2.5-coder emitted tool calls as plain text (0 of 6), and Ollama's default 4096-token context truncated OpenCode's ~10k-token prompt. Equally tuned, llama.cpp was within ~4% of Ollama, so Odin keeps Ollama.
- Catalog-based model recommendations that fit models to memory at the needed context ([MODEL_CATALOG.md](MODEL_CATALOG.md)).
- OpenCode runs at a 16k or 32k context through a derived Ollama model; `.odin/rules.md` is loaded; `~/.claude/CLAUDE.md` no longer leaks into sessions.
- `odin start --exec` exits non-zero when it cannot launch.
- `odin setup` shows the OpenCode install command instead of running it.
- Memory is off by default; MemPalace is opt-in.
- `odin doctor` warns about the Odin v1 `odin()` shell function.
- Accurate README with install, upgrade, uninstall and a list of files Odin creates; v1 docs moved to `legacy/README.md`.
- `odin opencode generate` exits non-zero when it refuses to overwrite a file.
- First-run smoke test (`scripts/smoke-test.sh`): the full CLI flow in an empty HOME without Ollama or OpenCode.

## Stage 1: reliable foundation (in progress)

- [ ] Merge `v2` into `main`, port `main`'s companion-temperature change into `legacy/setup.sh`, and freeze legacy.
- [ ] Tag `v0.2.0`; GitHub Actions for `go vet` and `go test`; release binaries with GoReleaser.
- [ ] Run `scripts/smoke-test.sh` in CI on a clean runner (the clean-install test).

## Stage 2: repeatable project workflow

- [ ] `odin init` creates `AGENTS.md` only when asked.
- [ ] Config ownership: `.odin/config.toml` holds shared intent; machine-specific values (model, Ollama URL) come from `~/.odin/config.toml`.
- [ ] Detect a hand-edited or stale generated config before regenerating.
- [ ] Show where each effective OpenCode setting comes from (global OpenCode config, Odin's config, a project `opencode.json`).
- [ ] Replace the hand-written config parser with a TOML library.

## Stage 3: companions as OpenCode agents (after the milestone)

- [ ] Map two or three companions to OpenCode agents with their own prompt, tools and permissions, using the existing `roles/` prompts.

## Open risks

- Qwen3.5/3.6 think by default, and OpenCode 1.19 offers no way to turn it off for Qwen models.
- OpenCode 1.x only gets triage upstream now; check compatibility with OpenCode 2.x.
- Model picks for 24 GB GPUs, Apple Silicon and CPU-only machines are unmeasured.
- macOS is untested.

## Deferred

- MemPalace / MCP memory integration beyond the current opt-in.
- Shared team workflows and onboarding.
- Shell prompt integration (Powerlevel10k, Starship).

## Non-goals

- Building an agent loop, a model runtime, a vector database, a RAG stack or a prompt engine.
