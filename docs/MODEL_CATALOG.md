# Odin model catalog

How `odin model recommend`, `odin setup` and `odin init` choose a local model, and where each number
comes from. Catalog data lives in `internal/models/catalog.go`; selection logic in
`internal/models/recommend.go`.

## How a recommendation is made

1. **Use case.** `agent` (default) is OpenCode with tool calls. OpenCode's system prompt and tool
   schemas take ~10.4k tokens before any user input, so agents need at least a 16k context (64k or
   32k when it fits). `chat` is coding Q&A without tools at an 8k context, and may use models whose tool
   calls are broken.
2. **Memory budget.** NVIDIA: VRAM minus 1 GiB. Apple Silicon: 65% of unified memory below 36 GB,
   75% from 36 GB (an assumption, not a documented limit). No GPU: 70% of RAM, and only models with at
   most 4B active parameters (CPU speed).
3. **Memory need.** Weights + KV cache + 0.6 GiB runtime overhead, where KV cache =
   bytes per token × context. Only full-attention layers grow with context, which is why the hybrid
   Qwen3.5/3.6 models need ~4-5x less KV memory than Qwen3.
4. **Pick.** The first catalog entry (best-first) that fits. If an agent pick's tool calls are only
   reported, the fallback is the best model Odin has verified, when one fits. Otherwise the fallback
   is the next smaller entry. The optional larger tier is a better entry that fits only with a q8_0 KV
   cache, which needs server-wide Ollama settings (`OLLAMA_FLASH_ATTENTION=1`,
   `OLLAMA_KV_CACHE_TYPE=q8_0`), so Odin never makes it the default.

Ollama's default context is 4096 on GPUs under ~24 GiB, which truncates OpenCode's prompt, and
its OpenAI-compatible endpoint cannot set `num_ctx` per request. So `odin opencode generate` points
OpenCode at a derived model such as `qwen3.5:9b-q4_K_M-ctx64k` (64k or 32k when the model fits,
otherwise 16k) with a matching OpenCode context limit, and `odin start --exec` creates or updates that
derived model through Ollama's `/api/create`. A long agent task outgrows 32k: OpenCode then compacts
the session into a summary, and the 9b model often lost the thread there (round 0 below).

## Sampling and thinking

Ollama's OpenAI-compatible endpoint sets temperature and top_p to 1.0 unless the request sends them,
and ignores top_k from requests. For a catalog entry with recommended sampling, Odin bakes all four
values into the derived model and has OpenCode send temperature and top_p (the model entry gets
`"temperature": true`, without which OpenCode never sends an agent's temperature).

`qwen3.5:9b-q4_K_M` uses Qwen's "precise coding" setting from its model card (temperature 0.6, top_p
0.95, top_k 20, presence_penalty 0; Ollama ships 1.0 and 1.5) with thinking off: the model entry sets
`"options": {"reasoningEffort": "none"}`, which Ollama maps to `think=false`. Thinking off passed
slightly more held-out runs and halved the median time (F1 against F2 below); the rerun of the final
build matched F1's 15/24 at a lower median. A `think` variant turns
it back on: `opencode run --variant think`, or ctrl+t in the OpenCode TUI. Its quality has not been
benchmarked.

## Catalog (verified 2026-10-28)

| Tag | Params (active) | Tag size | KV B/token | License | Open-source? | Tool calls |
|---|---|---|---|---|---|---|
| qwen3.6:27b-q4_K_M | 27B | 17.42 GB | 65,536 | Apache-2.0 | open-weight | reported |
| qwen3.6:35b-a3b-q4_K_M | 35B (3B) | 23.94 GB | 20,480 | Apache-2.0 | open-weight | reported |
| qwen3.5:35b-a3b-q4_K_M | 35B (3B) | 22.07 GB | 20,480 | Apache-2.0 | open-weight | reported |
| gpt-oss:20b | 20.9B (3.6B) | 13.79 GB | 24,576 | Apache-2.0 | open-weight | broken (chat only) |
| qwen3.5:9b-q4_K_M | 9B | 6.55 GB | 32,768 | Apache-2.0 | open-weight | **verified** |
| qwen3.5:4b-q4_K_M | 4B | 3.32 GB | 32,768 | Apache-2.0 | open-weight | reported |

None of these is open-source in the OSI sense (weights + training code + data). The only fully
open-source family found, AI2 OLMo 3/3.1, was too weak in agentic evaluations to rank.

## Measured by Odin (Stage 0, 2026-10-28)

RTX 5070 12 GB, 31 GB RAM, Ollama 0.40.12, OpenCode 1.19.4; task: add `Slugify` with a table test
in a small Go repo and make `go test` pass. One run per configuration unless noted.

| Model | Context | Result |
|---|---|---|
| qwen2.5-coder 3b / 7b (previous defaults) | 4k / 16k / 32k | 0/6: tool calls emitted as plain text, nothing executed |
| qwen2.5-coder:14b-instruct-q5_K_M (previous 12 GB pick) | 16k | spills to CPU (13%), ~10 tok/s |
| qwen3:8b | 16k / 32k | 0/2: tool loop works, task not completed |
| qwen3:14b | 16k / 32k | 2/4 across Ollama and llama.cpp configs; 18 tok/s with CPU spill at 16k |
| qwen3.5:9b-q4_K_M | 16k / 32k | **2/3** (passes in 81 s and 135 s); 100% GPU at 32k, 7.6 GiB, ~90 tok/s |

Memory estimates in `needGiB` matched nvidia-smi within 0.35 GiB for five model/context pairs
(`internal/models/memory_test.go`).

Known risks seen in measurement: qwen3.5:9b once produced a malformed XML tool call that Ollama
answered with HTTP 500 ("qwen3.5 tool call parsing failed").

## Measured by Odin (Exercism benchmark, 2026-10-28)

Same machine and versions. `scripts/bench.sh` runs medium Exercism Go exercises (difficulty 4-7,
exercism/go @ 6d9eed09): each run is a fresh repo with the stub, tests and instructions, `odin init`,
`odin opencode generate`, one `opencode run --agent baldur` (600 s limit), then `go test` against the
original tests. Dev set (tuning): poker, circular-buffer, save-the-cow, binary-search-tree. Held-out
set (reporting): robot-simulator, dot-dsl, meetup, tree-building, affine-cipher, paasio,
simple-linked-list, rectangles. Each change was tuned on dev x2 and reported on held-out x3 with no
task-specific hints. Reproduce the held-out table with:

```bash
odin start --exec   # once, so the context model exists with the current parameters; quit OpenCode
RUNS=3 AGENT=baldur scripts/bench.sh held-out
```

| Configuration (qwen3.5:9b-q4_K_M unless noted) | Dev | Held-out | Held-out median | Notes |
|---|---|---|---|---|
| Round 0: v0.2.0 (OpenCode's default agent, 32k) | 4/12 | 4/24 (17%) | 239 s | Stage 0 task 1/3; most failures follow compaction at ~24k |
| E: Baldur, 64k | 4/8 | 8/24 (33%) | 386.5 s | 8 timeouts; Stage 0 task 3/3 |
| F1: + Qwen coding sampling | 4/8 | 15/24 (62.5%) | 270 s | Stage 0 task 3/3 |
| **F2: + thinking off (shipped)** | 6/8 | **16/24 (66.7%)** | **121.5 s** | 7 timeouts; 1 false "tests pass" claim |
| F3: qwen3.6:35b-a3b (F2 settings, 59% on CPU) | 8/8 | stopped at 0/4 | – | swap grew 2.2 to 8.4 GiB on dev; held-out stopped at 6 GiB free RAM |
| F4: qwen3:14b, 16k, flash attention + q8_0 KV | 1/8 | not run | – | repeated identical-string edits; context hit the 12k compaction point |
| F5: F2 + OpenCode LSP (gopls) | 5/8 | not run | – | diagnostics reached the model; no gain |
| F6: F2 with Qwen's non-thinking sampling | 6/8 | 16/24 (66.7%) | 179.5 s | slower than F2 |
| Final build (F2 + `think` variant), rerun | – | 15/24 (62.5%) | 214 s | 8 timeouts; 0 false claims; Stage 0 task 3/3; Tyr changed no files in 3/3 |

GPU peaked at 80 C and CPU at 66 C; at 64k the 9b model ran 100% on GPU, with GPU memory use (desktop
included) at most 9.1 GiB of 12.
go vet passed on 13 of F2's 16 passing runs (12 of the rerun's 15), but only 6 were gofmt-clean (7 of
15; OpenCode's formatter is off by default). The loop stopped after two held-out iterations in a row without a 5-point gain.

The 9b model failed robot-simulator in every configuration, and the dev set barely separates
configurations (binary-search-tree and circular-buffer nearly always pass, poker nearly always
fails). Because F2 was chosen as the best of several held-out results, its 16/24 is optimistic; the rerun
of the final build passed 15/24.

## Sources

Opened directly while building this catalog, all verified 2026-10-28:

- Ollama registry manifests (tag sizes, including vision projectors):
  `https://registry.ollama.ai/v2/library/{qwen3.5/9b-q4_K_M, qwen3.5/4b-q4_K_M, qwen3.6/27b-q4_K_M,
  qwen3.6/35b-a3b-q4_K_M, qwen3.5/35b-a3b-q4_K_M, gpt-oss/20b, qwen3/14b, qwen3/8b}`
- Hugging Face `config.json` (layer types, KV heads, head dim):
  `https://huggingface.co/{Qwen/Qwen3.5-9B, Qwen/Qwen3.5-4B, Qwen/Qwen3.6-27B, Qwen/Qwen3.6-35B-A3B,
  Qwen/Qwen3.5-35B-A3B, openai/gpt-oss-20b}/resolve/main/config.json`
- https://github.com/ollama/ollama/issues/17638 (open: gpt-oss tool-call parsing)
- https://opencode.ai/docs/providers/ (Ollama `num_ctx` guidance), https://opencode.ai/docs/config/,
  https://opencode.ai/docs/rules/, https://opencode.ai/docs/permissions/

From the research pass of 2026-10-28 (sources opened by the research agent; Artificial Analysis and
Terminal-Bench figures were collected by one of its sub-agents and are marked [T] in `catalog.go`):

- Vendor model cards: https://huggingface.co/Qwen/Qwen3.6-27B, https://huggingface.co/Qwen/Qwen3.6-35B-A3B,
  https://huggingface.co/Qwen/Qwen3.5-35B-A3B, https://huggingface.co/Qwen/Qwen3.5-9B,
  https://huggingface.co/Qwen/Qwen3.5-4B, https://huggingface.co/openai/gpt-oss-20b
- Third-party: https://artificialanalysis.ai/leaderboards/models (Terminal-Bench 2.1, LiveCodeBench),
  https://gorilla.cs.berkeley.edu/data_overall.csv (BFCL),
  https://raw.githubusercontent.com/SWE-bench/swe-bench.github.io/master/data/leaderboards.json

## Not yet verified

- Tool calls and task quality for every "reported" entry; only qwen3.5:9b is measured among new picks.
- CPU-only and MoE speeds, and the Apple Silicon memory share.
- The `think` variant's quality, and OpenCode 2.x compatibility (OpenCode 1.x is now triage-only).

To update an entry, re-read its registry manifest and `config.json`, then rerun
`go test ./internal/models/`.
