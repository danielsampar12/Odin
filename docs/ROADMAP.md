# Odin v2 Roadmap

Updated 2026-10-28.

**Next milestone (still open): Odin works for someone who isn't its creator.** Three people on fresh machines (at least one macOS and one Linux) reach a successful, tool-using OpenCode edit through `odin start --exec`, using only the README, within 45 minutes and without help.

## Done

- Go CLI with `doctor`, `setup`, `init`, `start`, `model list/recommend/pull`, `opencode status/generate` and `memory status/doctor`.
- **Stage 0, core loop proven** on an RTX 5070 (12 GB): `qwen3.5:9b-q4_K_M` at a 16k–32k context completed the OpenCode test task in 2 of 3 runs. qwen2.5-coder emitted tool calls as plain text (0 of 6), and Ollama's default 4096-token context truncated OpenCode's ~10k-token prompt. Equally tuned, llama.cpp was within ~4% of Ollama, so Odin keeps Ollama.
- Catalog-based model recommendations that fit models to memory at the needed context ([MODEL_CATALOG.md](MODEL_CATALOG.md)).
- OpenCode runs at a 16k, 32k or 64k context through a derived Ollama model; `.odin/rules.md` is loaded; `~/.claude/CLAUDE.md` no longer leaks into sessions.
- `odin start --exec` exits non-zero when it cannot launch.
- `odin setup` shows the OpenCode install command instead of running it.
- Memory is off by default; MemPalace is opt-in.
- `odin doctor` warns about the Odin v1 `odin()` shell function.
- Accurate README with install, upgrade, uninstall and a list of files Odin creates; v1 docs moved to `legacy/README.md`.
- `odin opencode generate` exits non-zero when it refuses to overwrite a file.
- First-run smoke test (`scripts/smoke-test.sh`): the full CLI flow in an empty HOME without Ollama or OpenCode.
- CI (GitHub Actions): gofmt, `go vet`, `go test` and the smoke test on every push and pull request.
- `v2` merged into `main` (PR #4); `main`'s v1 temperature fix is ported into `legacy/setup.sh`, and legacy is frozen.
- Release workflow: pushing a `v*` tag builds Linux and macOS binaries (amd64, arm64) with checksums and publishes a GitHub release.
- **v0.2.0 released** (2026-10-28): https://github.com/danielsampar12/Odin/releases/tag/v0.2.0

## Stage 1: reliable foundation (done)

Next: run the milestone with three testers before starting Stage 3.

## Stage 2: repeatable project workflow (done)

- [x] `odin init` creates `AGENTS.md` only with `--agents-md`.
- [x] Config ownership: `.odin/config.toml` holds shared intent (companion, memory); the model and Ollama URL come only from `~/.odin/config.toml`, and stale project keys are ignored with a warning. `.odin/generated/` is machine-local; Odin prints the `.gitignore` line but never edits the file.
- [x] Drift: the generated config's header holds a content hash; `odin opencode status` reports up to date, out of date or edited by hand, and `generate` refuses to overwrite hand edits without `--force`.
- [x] Provenance: `odin opencode status` lists the OpenCode configs in OpenCode's merge order (global, `OPENCODE_CONFIG`, project `opencode.json[c]`, `.opencode/`, `OPENCODE_CONFIG_CONTENT`; verified with `opencode debug config` on 1.19.4), the instruction files, and warns when a later config overrides the model.
- [x] TOML library (BurntSushi/toml) instead of the hand-written parser; invalid TOML now fails with the file name.
- [x] Launches also set `OPENCODE_DISABLE_EXTERNAL_SKILLS=1`: OpenCode otherwise lists every skill under `~/.claude/skills` and `~/.agents/skills` (~14k characters on the test machine) in each session's prompt.
- [x] `odin model pull` and `odin model list` exit non-zero when Ollama is unreachable or a pull fails; pulling an installed model is a no-op.

## Stage 3: companions as OpenCode agents (done, ahead of the milestone)

The plan put Stage 3 after the three-tester milestone; the maintainer chose to build it first. The milestone stays open.

- [x] Baldur (default) and Tyr (reviewer) are OpenCode primary agents in the generated config, with prompts from `roles/*.md` embedded in the binary; `[companion] default` sets OpenCode's `default_agent`.
- [x] Tyr cannot edit or write files, run shell commands, fetch URLs or start subagents; it reads through OpenCode's read, grep and glob tools. (An allowlist such as `git log*` or `ls*` still lets a shell write files.) No companion may start subagents (no orchestration).
- [x] `odin companion list` and `odin companion show <name>`.
- [x] Hephaestus, Freya, Thor and Loki are dropped from Odin v2 until they have agent definitions; a config that names one falls back to Baldur with a warning.
- [x] `roles/baldur.md` is rewritten for agent work, because an agent prompt replaces OpenCode's built-in prompt. Odin v1 (`legacy/`) reads the same file.

## Benchmark and tuning (done)

- [x] `scripts/bench.sh`: medium Exercism Go exercises, 4 dev and 8 held-out, each in a fresh repo through `odin init`, `odin opencode generate` and one `opencode run`, with telemetry guardrails (`scripts/telemetry.sh`), code-quality checks (`scripts/bench-quality.sh`) and the Tyr no-changes check (`scripts/tyr-check.sh`).
- [x] Held-out pass@1 for qwen3.5:9b went from 17% (v0.2.0, 32k) to 71% and then 62.5% on a rerun, at a median under 2 minutes: Baldur at 64k, Qwen's coding sampling, thinking off (with a `think` variant to turn it back on) and a smaller output limit. Every iteration is in [MODEL_CATALOG.md](MODEL_CATALOG.md).

## Open risks

- Thinking off is the 9b default; the `think` variant is unbenchmarked, and one held-out run claimed passing tests that hung.
- An agent's own runaway `go test` can drain host RAM (seen once: 0.7 GiB free, 15 GiB of swap); OpenCode puts no memory limit on shell commands.
- The benchmark's dev set barely separates configurations, and the shipped config was the best of several held-out results, so its score is optimistic: F2 and F7 both reran at 15/24, and F7's edge over F2 is speed, not score. F7's rounds ran on a newer kernel than the rest.
- qwen3.6:35b-a3b solved every dev task on a 12 GB GPU but spilled to RAM and grew swap by 6 GiB next to a desktop on 32 GB. The CPU-only pick for 32 GB machines, qwen3.5:35b-a3b, is the same size and may swap the same way (unmeasured).
- The GPU driver (595.99.02) crashed once during the benchmark (Xid 62) and needed a reboot; until then Ollama silently ran the model on CPU.
- OpenCode 1.x only gets triage upstream now; check compatibility with OpenCode 2.x.
- Model picks for 24 GB GPUs, Apple Silicon and CPU-only machines are unmeasured.
- macOS is untested.

## Deferred

- MemPalace / MCP memory integration beyond the current opt-in.
- Shared team workflows and onboarding.
- Shell prompt integration (Powerlevel10k, Starship).

## Non-goals

- Building an agent loop, a model runtime, a vector database, a RAG stack or a prompt engine.
