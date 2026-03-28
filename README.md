<div align="center">

# ⚡ Hermes

<p>
  <img src="https://img.shields.io/badge/offline-100%25-brightgreen?style=flat-square" />
  <img src="https://img.shields.io/badge/platform-macOS%20%7C%20Linux-lightgrey?style=flat-square" />
  <img src="https://img.shields.io/badge/powered%20by-Ollama-black?style=flat-square" />
  <img src="https://img.shields.io/badge/license-MIT-blue?style=flat-square" />
</p>

*A local AI coding assistant. No cloud. No tokens. Just you, your machine,*
*and a god who literally invented writing.*

</div>

---

Hermes sets up a fully offline AI companion that lives in your terminal. Your code never leaves your machine. No API keys, no subscriptions, no sending your half-finished startup idea to a server farm somewhere.

Run the setup once, pick your companion, and get to work.

## Requirements

- **macOS or Linux**
- [Homebrew](https://brew.sh) — required on macOS, optional on Linux

## Install

```bash
git clone https://github.com/danielsampar12/Hermes.git ~/ai/hermes
cd ~/ai/hermes
chmod +x setup.sh
./setup.sh
```

> [!NOTE]
> Prefer SSH? Use `git@github.com:danielsampar12/Hermes.git` instead.

The script will introduce itself, ask for a name, let you pick your companion, detect your hardware, pull the right model, and configure everything. Reload your shell when it's done:

```bash
source ~/.zshrc  # or ~/.bashrc
```

---

## Contents

- [Companions](#companions)
- [Usage](#usage)
- [Configuration](#configuration)
- [Hardware & models](#hardware--models)
- [Storage](#storage)

---

## Companions

You don't get a generic chatbot. You pick who rides with you.

| | Chiron | Ares |
|---|---|---|
| **Role** | Pair programmer | Implementer |
| **Style** | Asks before acting | Acts, then mentions it |
| **Best for** | Learning, complex decisions, architecture | Getting things done fast |
| **Will push back?** | Yes, always | Only if it's truly wrong |
| **Mythology** | Immortal centaur. Trained Achilles, Jason, Asclepius. Chose death to free Prometheus. | God of war. Passionate, fierce. His peers trapped him in a bronze jar once. He did not enjoy that. |

> [!TIP]
> You can override your default companion per session — see [Usage](#usage).

---

## Usage

```bash
hermes start              # wake up
hermes stop               # rest
hermes my-project         # start or resume a session (uses your default companion)
hermes new                # fresh unnamed session
hermes list               # see all sessions
hermes chiron my-project  # summon Chiron for this session
hermes ares my-project    # summon Ares for this session
```

Sessions are saved automatically and resume with full history. You only explain your project once — Chiron will remember.

---

## Configuration

**Change a companion's personality** — edit the role file and copy it over:

```bash
vim ~/ai/hermes/roles/chiron.md
cp ~/ai/hermes/roles/chiron.md ~/.config/aichat/roles/chiron.md
```

**Switch your default companion or model** — just re-run setup:

```bash
./setup.sh
```

**aichat config** lives at `~/.config/aichat/config.yaml` if you want to tweak anything manually.

---

## Hardware & models

<details>
<summary>Click to expand — model recommendations by hardware</summary>

The setup script auto-detects your hardware and recommends the best model. Here's the logic:

| Hardware | Recommended | Safe fallback |
|---|---|---|
| NVIDIA 24GB+ VRAM | `qwen3-coder:30b` | `qwen2.5-coder:32b` |
| NVIDIA 12–16GB VRAM | `qwen2.5-coder:14b-instruct-q5_K_M` | `q4_K_M` |
| NVIDIA 8GB VRAM | `qwen2.5-coder:7b` | `qwen2.5-coder:7b` |
| NVIDIA <8GB VRAM | `qwen2.5-coder:3b` | `qwen2.5-coder:3b` |
| Apple Silicon 48GB+ | `qwen3-coder:30b` | `qwen2.5-coder:32b` |
| Apple Silicon 32GB | `qwen3-coder:30b` | `qwen2.5-coder:14b-instruct-q5_K_M` |
| Apple Silicon 16GB | `qwen2.5-coder:14b-instruct-q5_K_M` | `q4_K_M` |
| Apple Silicon 8GB | `qwen2.5-coder:7b` | `qwen2.5-coder:3b` |
| Intel Mac / no GPU | `qwen2.5-coder:7b` | `qwen2.5-coder:3b` |

**Why these models:**
- `qwen3-coder:30b` — best open-source coding model as of 2025 (70.6% SWE-Bench), 256K context, fits in ~19GB
- `qwen2.5-coder:14b` — best-in-class for 8–16GB VRAM, battle-tested, fast
- `qwen2.5-coder:7b` — reliable for tighter hardware, still very capable

> [!NOTE]
> Apple Silicon uses unified memory — your RAM is your VRAM. A 48GB M4 handles 30B models with room to spare.

</details>

---

## Storage

| What | Path | Size |
|---|---|---|
| Models | `~/.ollama/models/` | ~10–20GB per model |
| Sessions | `~/ai/sessions/` | Tiny (plain text) |
| aichat config | `~/.config/aichat/` | Tiny |

```bash
ollama list              # see downloaded models
ollama rm <model-name>   # free up space
```

---

<div align="center">
<sub>Built with Ollama + aichat. Hermes takes no responsibility for code shipped at 2am.</sub>
</div>
