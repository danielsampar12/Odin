<div align="center">

# Odin

<p>
  <img src="https://img.shields.io/badge/local--first-100%25-brightgreen?style=flat-square" />
  <img src="https://img.shields.io/badge/platform-Linux%20%7C%20macOS-lightgrey?style=flat-square" />
  <img src="https://img.shields.io/badge/runtime-Ollama-black?style=flat-square" />
  <img src="https://img.shields.io/badge/agent-OpenCode-blue?style=flat-square" />
</p>

*A local AI coding companion. No cloud. No tokens. Just you, your machine,*
*and a god who hung from a tree for nine days to read the runes.*
*He's seen worse codebases*

<p>
  <a href="#requirements">Requirements</a> •
  <a href="#install">Install</a> •
  <a href="#quickstart">Quickstart</a> •
  <a href="#commands">Commands</a> •
  <a href="#what-odin-changes-on-your-machine">Files</a> •
  <a href="#models">Models</a> •
  <a href="#uninstall">Uninstall</a>
</p>

</div>

Odin sets up and launches a local AI coding stack for a repository. [Ollama](https://ollama.com) runs the model, [OpenCode](https://opencode.ai) is the coding agent, and Odin handles the parts in between: it checks your machine, picks a model that fits it, writes project-local configuration, and launches OpenCode with a context large enough for OpenCode's prompt. Your code never leaves your machine.

> [!NOTE]
> This is **Odin v2**, a Go CLI under active development. The original shell setup (Odin v1) is preserved in [`legacy/`](legacy/README.md).

## Requirements

- **Linux.** Tested on Ubuntu with an NVIDIA GPU. macOS should work but is untested, and `odin setup` only automates steps on Linux.
- **[Go](https://go.dev/dl/) 1.25 or newer**, only if you install with Go or build from source.
- **[Ollama](https://ollama.com/download)** and **[OpenCode](https://opencode.ai/docs/)** 1.x (tested with 1.19.4). Odin shows the install commands but does not install either for you.
- **Disk space for one model:** 3–24 GB depending on your hardware (see [Models](#models)).

## Install

**Release binary** (Linux and macOS, amd64 and arm64): download `odin_<version>_<os>_<arch>.tar.gz` from [Releases](https://github.com/danielsampar12/Odin/releases), check it against `checksums.txt`, and put `odin` on your `PATH`:

```bash
tar -xzf odin_v0.2.0_linux_amd64.tar.gz
mkdir -p ~/.local/bin && mv odin ~/.local/bin/
```

**With Go:**

```bash
go install github.com/danielsampar12/odin/cmd/odin@latest
```

**From source:**

```bash
git clone https://github.com/danielsampar12/Odin.git
cd Odin
go install ./cmd/odin
```

`go install` puts the binary in `$(go env GOPATH)/bin`, usually `~/go/bin`. If that directory is not on your `PATH` yet, add it in `~/.zshrc` or `~/.bashrc` and open a new shell:

```bash
export PATH="$(go env GOPATH)/bin:$PATH"
```

Then check:

```bash
odin version
```

**Upgrade:** download the newer release, run the `go install …@latest` command again, or `git pull && go install ./cmd/odin` in your clone.

> [!WARNING]
> If you used Odin v1, its setup added an `odin()` function to `~/.zshrc` or `~/.bashrc`. That function hides the v2 binary. Run `command odin doctor` to bypass it: doctor names the file, and you can delete the `odin() { ... }` block and open a new shell.

## Quickstart

Once per machine:

```bash
odin doctor            # what Odin sees: OS, GPU, RAM, Ollama, OpenCode
odin setup --dry-run   # the setup plan, without changing anything
odin setup             # writes ~/.odin/config.toml and runs the automatic steps after you confirm
```

Then, inside a repository:

```bash
odin init                            # .odin/config.toml and .odin/rules.md (--agents-md adds AGENTS.md)
odin model recommend                 # the model that fits this machine
odin model pull <recommended-model>  # skip if setup already pulled it
odin opencode generate               # .odin/generated/opencode.jsonc for this machine
odin start --exec                    # validate everything, then launch OpenCode
```

## Commands

| Command | What it does |
|---|---|
| `odin doctor` | Read-only diagnostics for this machine and project |
| `odin setup [--dry-run] [--yes]` | Plans machine setup. Writes `~/.odin/config.toml` and pulls the recommended model; installing Ollama and OpenCode stays a manual step it describes |
| `odin model recommend [--use-case agent\|chat]` | Recommends a model for this machine and the context size to run it at |
| `odin model list` | Lists models installed in Ollama |
| `odin model pull <name> [--yes]` | Pulls a model into Ollama |
| `odin init [--agents-md]` | Creates the project files; never overwrites existing ones |
| `odin opencode generate [--force] [--with-memory]` | Writes the machine-local OpenCode config; refuses to replace a file Odin did not generate, or one edited by hand, unless `--force` |
| `odin opencode status` | Shows whether the generated config is up to date, out of date or edited by hand, and where each OpenCode setting comes from |
| `odin start [--exec]` | Validates the stack. With `--exec`, creates the context model if needed and launches OpenCode; exits non-zero if it cannot |
| `odin companion list`, `odin companion show <name>` | Lists the companions and shows each one's OpenCode agent config and prompt |
| `odin memory status`, `odin memory doctor` | Inspects the optional MemPalace integration |
| `odin version` | Prints the version |

## What Odin changes on your machine

| Path | Created by | Notes |
|---|---|---|
| `~/.odin/config.toml` | `odin setup` | Machine settings: model, Ollama URL, default companion |
| `.odin/config.toml` | `odin init` | Shared project settings (companion, memory); safe to commit. The model and Ollama URL are machine-specific and only come from `~/.odin/config.toml` |
| `.odin/rules.md` | `odin init` | Project instructions that OpenCode loads; safe to commit |
| `AGENTS.md` | `odin init --agents-md` | Only with the flag, and only when the repository has none |
| `.odin/generated/opencode.jsonc` | `odin opencode generate` | Machine-specific: keep it out of version control (Odin prints the `.gitignore` line but never edits `.gitignore`). Its header holds a hash, so Odin can tell a stale file from a hand-edited one |
| Ollama models | `odin setup`, `odin model pull` | The model weights (several GB each) |
| Ollama models named `<model>-ctx16k` or `<model>-ctx32k` | `odin start --exec` | The same weights with a larger context; takes almost no extra disk |

Odin does not edit your shell startup files, your `.gitignore`, your global OpenCode config, or Ollama's server settings. When it launches OpenCode, it sets `OPENCODE_DISABLE_CLAUDE_CODE=1` and `OPENCODE_DISABLE_EXTERNAL_SKILLS=1`, so your global `~/.claude/CLAUDE.md` and the skills in `~/.claude/skills` and `~/.agents/skills` do not leak into the project; skills in the project's `.opencode/skills/` still load.

Put your own project-level OpenCode settings (permissions, extra instructions) in `opencode.json` at the repository root: OpenCode applies it after Odin's generated config. If it sets `model`, it overrides Odin's model, and `odin opencode status` warns about it.

**Upgrading from v0.2:** Odin now ignores `[model]` and `[runtime] base_url` in `.odin/config.toml` (it warns until you delete them), and the first `odin start --exec` asks you to re-run `odin opencode generate`.

## Models

OpenCode's prompt is about 10k tokens before you type anything, and Ollama's default context (4096 tokens on most GPUs) cuts it off. So `odin model recommend` picks the best model that fits your memory with at least 16k tokens of context, and Odin runs it at that size.

| Machine | Recommended for agent use |
|---|---|
| No GPU, 16 GB RAM | `qwen3.5:4b-q4_K_M` |
| No GPU, 32 GB RAM | `qwen3.5:35b-a3b-q4_K_M` |
| NVIDIA 8 GB | `qwen3.5:4b-q4_K_M` |
| NVIDIA 12–16 GB, Apple Silicon 16–24 GB | `qwen3.5:9b-q4_K_M` |
| NVIDIA 24 GB, Apple Silicon 32 GB and up | `qwen3.6:27b-q4_K_M` |

Only `qwen3.5:9b-q4_K_M` has been measured end to end with OpenCode so far; the other picks rest on published benchmarks. Method, measurements and sources: [docs/MODEL_CATALOG.md](docs/MODEL_CATALOG.md).

## Memory

Memory is off by default. MemPalace support is optional: set `provider = "mempalace"` under `[memory]` in `.odin/config.toml`, install MemPalace yourself, then run `odin opencode generate --with-memory`.

## Companions

Companions are OpenCode agents that `odin opencode generate` writes into the project's generated config, each with its own prompt (from [`roles/`](roles/)) and permissions:

| Companion | What it does |
|---|---|
| **Baldur** (default) | Pair programmer: makes the changes you ask for, runs the tests, and explains tradeoffs |
| **Tyr** | Reviewer: reads code and diffs, runs `go test`/`go vet` and reports problems. It cannot edit or write files, fetch URLs or start subagents, and its shell is limited to read-only commands |

OpenCode starts on the companion set by `[companion] default` in `.odin/config.toml` (or `~/.odin/config.toml`); press Tab in OpenCode to switch, or use `opencode run --agent tyr`. `odin companion list` shows them and `odin companion show <name>` prints the exact agent config and prompt. A companion's prompt replaces OpenCode's built-in prompt for that agent; OpenCode's own `build` and `plan` agents stay available. After upgrading Odin, run `odin opencode generate` again: a changed prompt makes the generated config out of date.

## Uninstall

```bash
rm "$(go env GOPATH)/bin/odin"
rm -rf ~/.odin
```

In each project, delete `.odin/`, and `AGENTS.md` if Odin created it. `ollama list` shows the `-ctx16k`/`-ctx32k` models Odin created; remove them with `ollama rm <name>`. Ollama, OpenCode and your models are yours to keep or uninstall.

## Odin v1

The original shell setup, with companions, sessions and RAG through aichat, is documented in [legacy/README.md](legacy/README.md). `./setup.sh` still forwards to it.

## Roadmap

See [docs/ROADMAP.md](docs/ROADMAP.md) for what is done and what comes next, and [docs/VISION.md](docs/VISION.md) for the direction.

---

<div align="center">
<sub>Built with Ollama + OpenCode. Odin takes no responsibility for code shipped at 2am.</sub>
</div>
