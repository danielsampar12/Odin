package models

import (
	"fmt"
	"math"

	"github.com/danielsampar12/odin/internal/system"
)

// UseCase is what the model will be used for. It sets the context size and whether tool calls must work.
type UseCase string

const (
	UseCaseAgent UseCase = "agent" // OpenCode-style agentic coding: needs working tool calls
	UseCaseChat  UseCase = "chat"  // coding questions and explanations without tool calls
)

// OpenCode's system prompt and tool schemas take ~10.4k tokens before any user input, so agents need
// at least a 16k context; Ollama's default (4096 on a 12 GB GPU) silently truncates it.
const (
	agentContext = 16384
	chatContext  = 8192
)

// roomierAgentContexts are tried largest first. A long agent task outgrows 32k: OpenCode then compacts
// the session into a summary, and a small model often loses the thread there (Odin's Exercism benchmark).
var roomierAgentContexts = []int{65536, 32768}

// agentContextFor is the largest agent context m fits in budget, or the 16k minimum.
func agentContextFor(m Model, budget float64) int {
	for _, ctx := range roomierAgentContexts {
		if needGiB(m, ctx, false) <= budget {
			return ctx
		}
	}
	return agentContext
}

type Recommendation struct {
	Recommended          string
	Fallback             string
	OptionalLarger       string
	OptionalLargerReason string
	Context              int // num_ctx to run Recommended with
	Reason               string
}

// RecommendCodingModel recommends a model for agentic coding with OpenCode, Odin's default workflow.
func RecommendCodingModel(ramGB int, gpu system.GPUInfo) Recommendation {
	return Recommend(ramGB, gpu, UseCaseAgent)
}

// Recommend picks the best catalog model that fits this machine at the context the use case needs.
// Catalog order is the quality ranking. When an agent default is only reported to support tools,
// the fallback is the best model Odin has verified, because Ollama's "tools" capability flag alone
// has proven unreliable (qwen2.5-coder advertises it but emits calls as plain text).
func Recommend(ramGB int, gpu system.GPUInfo, useCase UseCase) Recommendation {
	budget, machine := memoryBudget(ramGB, gpu)
	ctx := chatContext // any use case other than agent is treated as chat
	if useCase == UseCaseAgent {
		ctx = agentContext
	}
	cpuOnly := !gpu.Detected && !gpu.AppleSilicon

	// TestCatalogInvariants guarantees at least one suitable entry per use case.
	var usable []Model
	for _, m := range Catalog {
		if suitable(m, useCase, cpuOnly) {
			usable = append(usable, m)
		}
	}

	best := -1
	for i, m := range usable {
		if needGiB(m, ctx, false) <= budget {
			best = i
			break
		}
	}
	if best < 0 {
		smallest := usable[len(usable)-1]
		return Recommendation{
			Recommended: smallest.Tag,
			Fallback:    smallest.Tag,
			Context:     ctx,
			Reason: fmt.Sprintf("%s. No catalog model fits comfortably with %d tokens of context (%s needs about %.1f GiB), so Odin is suggesting the smallest one; expect slow or partial-CPU inference.",
				machine, ctx, smallest.Tag, needGiB(smallest, ctx, false)),
		}
	}

	pick := usable[best]
	rec := Recommendation{Recommended: pick.Tag, Fallback: pick.Tag, Context: ctx}
	if useCase == UseCaseAgent {
		rec.Context = agentContextFor(pick, budget)
	}
	unverifiedAgent := useCase == UseCaseAgent && pick.ToolCalls != ToolCallsVerified
	fallback := pick
	if unverifiedAgent {
		for _, m := range usable[best+1:] {
			if m.ToolCalls == ToolCallsVerified && needGiB(m, ctx, false) <= budget {
				fallback = m
				break
			}
		}
	}
	if fallback.Tag == pick.Tag {
		for _, m := range usable[best+1:] {
			if needGiB(m, ctx, false) < needGiB(pick, ctx, false) {
				fallback = m
				break
			}
		}
	}
	rec.Fallback = fallback.Tag
	for _, m := range usable[:best] {
		if needGiB(m, ctx, true) <= budget {
			rec.OptionalLarger = m.Tag
			rec.OptionalLargerReason = fmt.Sprintf("Fits at the %d-token minimum context only with a q8_0 KV cache (about %.1f GiB): set OLLAMA_FLASH_ATTENTION=1 and OLLAMA_KV_CACHE_TYPE=q8_0 on the Ollama server, which affects every model it serves.", ctx, needGiB(m, ctx, true))
			break
		}
	}

	rec.Reason = fmt.Sprintf("%s. %s fits with %d tokens of context (about %.1f GiB). %s",
		machine, pick.Tag, rec.Context, needGiB(pick, rec.Context, false), pick.Evidence)
	switch {
	case unverifiedAgent && fallback.ToolCalls == ToolCallsVerified:
		rec.Reason += fmt.Sprintf(" Odin has not verified its tool calls with OpenCode yet; %s is the measured fallback.", fallback.Tag)
	case unverifiedAgent:
		rec.Reason += " Odin has not verified its tool calls with OpenCode yet, and no verified model is suitable for this machine at default settings."
	}
	if cpuOnly && useCase == UseCaseAgent {
		rec.Reason += " Without a GPU, expect minutes per agent turn: the ~10k-token prompt is processed on the CPU."
	}
	return rec
}

// ContextFor returns the context Odin runs an agent model with on this machine: the largest of 64k and
// 32k that fits, otherwise the 16k minimum OpenCode needs. Models outside the catalog get the minimum.
func ContextFor(tag string, ramGB int, gpu system.GPUInfo) int {
	budget, _ := memoryBudget(ramGB, gpu)
	for _, m := range Catalog {
		if m.Tag == tag {
			return agentContextFor(m, budget)
		}
	}
	return agentContext
}

// suitable filters the catalog by use case and by speed on CPU-only machines.
func suitable(m Model, useCase UseCase, cpuOnly bool) bool {
	if useCase == UseCaseAgent && m.ToolCalls == ToolCallsBroken {
		return false
	}
	// shortcut: a fixed active-parameter cap stands in for real CPU speed estimates; replace it with measured tok/s if CPU users report problems.
	return !cpuOnly || m.ActiveParamsB <= 4
}

const gib = 1 << 30

// runtimeOverheadGiB covers the CUDA/Metal context and compute buffers; measured ~0.6 GiB on an RTX 5070.
const runtimeOverheadGiB = 0.6

// needGiB estimates memory for a model at a context size: weights + KV cache + runtime overhead.
// It matched every fit and CPU spill measured in Odin's Stage 0 runs (Ollama 0.40, llama.cpp).
func needGiB(m Model, ctx int, q8KV bool) float64 {
	kv := float64(m.KVBytesPerToken) * float64(ctx)
	if q8KV {
		kv /= 2
	}
	return m.WeightsGB*1e9/gib + kv/gib + runtimeOverheadGiB
}

// memoryBudget returns the memory a model can use without spilling, and a description of the machine.
func memoryBudget(ramGB int, gpu system.GPUInfo) (float64, string) {
	switch {
	case gpu.AppleSilicon:
		// macOS lets the GPU use roughly 2/3 of unified memory on smaller machines and ~3/4 on larger ones.
		share := 0.65
		if ramGB >= 36 {
			share = 0.75
		}
		budget := float64(ramGB) * share
		return budget, fmt.Sprintf("Detected Apple Silicon with about %dGB of unified memory (about %.0f GiB usable by the GPU)", ramGB, math.Floor(budget))
	case gpu.Detected:
		// Reserve ~1 GiB for the display and driver; VRAMGB is rounded up from the reported MiB.
		budget := float64(gpu.VRAMGB) - 1
		return budget, fmt.Sprintf("Detected an NVIDIA GPU with about %dGB VRAM (about %.0f GiB usable)", gpu.VRAMGB, math.Floor(budget))
	default:
		// Leave ~30% of RAM for the OS, editor and browser.
		budget := float64(ramGB) * 0.7
		return budget, fmt.Sprintf("No dedicated GPU was detected; with %dGB RAM, Odin budgets about %.0f GiB for the model", ramGB, math.Floor(budget))
	}
}
