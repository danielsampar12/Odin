<div align="center">

# Odin

<p>
  <img src="https://img.shields.io/badge/offline-100%25-brightgreen?style=flat-square" />
  <img src="https://img.shields.io/badge/platform-macOS%20%7C%20Linux-lightgrey?style=flat-square" />
  <img src="https://img.shields.io/badge/powered%20by-Ollama-black?style=flat-square" />
  <img src="https://img.shields.io/badge/license-MIT-blue?style=flat-square" />
</p>

*A local AI coding companion. No cloud. No tokens. Just you, your machine,*
*and a god who hung from a tree for nine days to read the runes.*
*He's seen worse codebases*

<p>
  <a href="#requirements">Requirements</a> •
  <a href="#install">Install</a> •
  <a href="#companions">Companions</a> •
  <a href="#usage">Usage</a> •
  <a href="#configuration">Configuration</a> •
  <a href="#hardware--models">Hardware</a> •
  <a href="#uninstall">Uninstall</a> •
  <a href="#roadmap">Roadmap</a>
</p>

</div>

Odin sets up a fully offline AI companion that lives in your terminal. Your code never leaves your machine. No API keys, no subscriptions, no sending your half-finished startup idea to a server farm somewhere.

Run the setup once, pick your companion, and get to work.

## Requirements

- **macOS or Linux**
- [Homebrew](https://brew.sh) — required on macOS. On Linux the script handles everything without it.

## Install
> [!IMPORTANT]
> The model download can be **10–20GB** depending on your hardware. Make sure you're on a good connection and have time — this is a one-time step.
```bash
git clone https://github.com/danielsampar12/Odin.git ~/ai/odin
cd ~/ai/odin
chmod +x setup.sh
./setup.sh
```

The script will introduce itself, ask for a name, let you pick your companion, detect your hardware, pull the right model, and configure everything. Reload your shell when it's done:

```bash
source ~/.zshrc  # or ~/.bashrc
```

## Companions

You don't get a generic chatbot. You pick who rides with you into battle.

### ✨ Baldur — pair programmer

Most beloved of all gods in Asgard. Patient, wise, trusted by everyone — nobody could hate Baldur, and that patience shows in how he works.

- **Style:** asks before acting, suggests an approach and waits for your go-ahead
- **Best for:** complex decisions, architecture, learning, understanding the codebase
- **Will push back?** yes — if you're heading toward a bad pattern he'll tell you why before touching anything
- **Explains:** key decisions as he writes, so you never lose track of what's happening

### ⚖️ Tyr — architect / reviewer

God of law and justice. He sacrificed his hand to bind the wolf Fenrir — he knew exactly what it would cost and paid it anyway. That's the mindset of someone who reviews code and says "this will hurt later."

- **Style:** reviews from the outside in — structure, abstractions, and whether you're solving the right problem
- **Best for:** code reviews, architecture decisions, spotting technical debt before it bites
- **Will push back?** yes, and he'll tell you exactly why it will hurt later — with an alternative
- **Explains:** the systemic risk, not just the immediate bug

### ⚡ Thor — implementer

God of thunder. Direct, powerful, and not exactly known for overthinking. The hammer doesn't ask questions.

- **Style:** acts first, mentions it if something was off
- **Best for:** getting things done fast, clear and well-defined tasks
- **Will push back?** only if something is genuinely wrong — one question, no more
- **Explains:** only when the why isn't obvious from the code itself

### 🃏 Loki — chaos agent

> [!WARNING]
> Not recommended for production. Review everything he gives you. You have been warned.

Shapeshifter. Trickster. The god who always finds a third option nobody else considered. He will solve your problem — just not the way you expected.

- **Style:** ignores conventions when a better solution exists outside them
- **Best for:** creative problems, breaking out of tunnel vision, exploring unconventional approaches
- **Will push back?** he'll reframe the entire question
- **Explains:** why his approach works and what could go wrong — he's chaotic, not irresponsible

> [!TIP]
> You can override your default companion per session.

## Usage

> [!NOTE]
> Examples below use `odin`, but the setup script lets you pick any name.
> Odin has over two hundred names across the nine realms.
> Whatever you chose, use that instead.

```bash
odin start          # awaken
odin stop           # return to Asgard

odin my-project     # start or resume a session (uses your default companion)
odin new            # fresh unnamed session
odin list           # see all sessions

odin baldur my-project  # summon Baldur for this session
odin tyr my-project     # summon Tyr for this session
odin thor my-project    # summon Thor for this session
odin loki my-project    # summon Loki (you were warned)

odin add                # forge a new custom companion
odin remove my-project  # remove a session

odin index my-project   # index project files for RAG (run once per project)
# next time you start a session named 'my-project', RAG is detected automatically
```

Sessions are saved automatically and resume with full history. You only explain your project once.

## Configuration

**Change a companion's personality** — edit the role file and copy it over:

```bash
vim ~/ai/odin/roles/baldur.md
cp ~/ai/odin/roles/baldur.md ~/.config/aichat/roles/baldur.md
# on macOS: ~/Library/Application\ Support/aichat/roles/baldur.md
```

**Switch your default companion or model** — just re-run setup:

```bash
./setup.sh
```

**Add a custom companion** — run `odin add` and follow the prompts. You can either provide a path to an existing `.md` file or paste the role definition directly. The companion is installed alongside the built-ins and works the same way:

```bash
odin add              # interactive — name it, then provide a file path or paste text
odin freya my-project # once added, summon it like any other companion
```

**aichat config** lives at `~/.config/aichat/config.yaml` (Linux) or `~/Library/Application Support/aichat/config.yaml` (macOS).

## Uninstall

```bash
cd ~/ai/odin
./uninstall.sh
```

Removes Ollama, aichat, configs, and the shell function. Asks before touching models and sessions since those are large or personal.

## Hardware & models

<details>
<summary>Model recommendations by hardware</summary>

The setup script auto-detects your hardware and recommends the best model.

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

## Storage

| What | Path | Size |
|---|---|---|
| Models | `~/.ollama/models/` | ~10–20GB per model |
| Sessions | `~/ai/sessions/` | Tiny (plain text) |
| aichat config | `~/.config/aichat/` (Linux) or `~/Library/Application Support/aichat/` (macOS) | Tiny |

```bash
ollama list              # see downloaded models
ollama rm <model-name>   # free up space
```

## Roadmap

- [x] RAG support — `odin index my-project` indexes your codebase; detected automatically on session start
- [ ] Smart VRAM fallback — detect OOM at runtime and automatically reload the session with the safe fallback model
- [ ] Hardware-aware `num_ctx` — scale context window size based on detected VRAM/RAM tier (needs benchmarking)
- [ ] Tune Thor, Tyr and Loki roles
- [ ] Web search support — fetch docs and inject them into context on demand
- [ ] Anything else
- [ ] Windows support

---

<div align="center">
<sub>Built with Ollama + aichat. Odin takes no responsibility for code shipped at 2am.</sub>
</div>
