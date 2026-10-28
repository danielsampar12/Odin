# Odin model catalog

How `odin model recommend`, `odin setup` and `odin init` choose a local model, and where each number
comes from. Catalog data lives in `internal/models/catalog.go`; selection logic in
`internal/models/recommend.go`.

## How a recommendation is made

1. **Use case.** `agent` (default) is OpenCode with tool calls. OpenCode's system prompt and tool
   schemas take ~10.4k tokens before any user input, so agents need at least a 16k context (32k when
   it fits). `chat` is coding Q&A without tools at an 8k context, and may use models whose tool
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
OpenCode at a derived model such as `qwen3.5:9b-q4_K_M-ctx32k` (32k when the model fits, otherwise
16k) with a matching OpenCode context limit, and `odin start --exec` creates that derived model
through Ollama's `/api/create` if it is missing.

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
answered with HTTP 500 ("qwen3.5 tool call parsing failed"); Qwen3.5/3.6 think by default and
OpenCode 1.19 exposes no reasoning control for Qwen models.

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
- Thinking on vs off with tools at 32k, and OpenCode 2.x compatibility (OpenCode 1.x is now triage-only).

To update an entry, re-read its registry manifest and `config.json`, then rerun
`go test ./internal/models/`.
