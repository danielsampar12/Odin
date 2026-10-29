package models

// ToolCalls records how a model's tool calls behave through Ollama with an agent such as OpenCode.
type ToolCalls int

const (
	ToolCallsBroken   ToolCalls = iota // calls fail through Ollama: measured by Odin or an open upstream bug
	ToolCallsReported                  // advertised by the model card and Ollama, not yet measured by Odin
	ToolCallsVerified                  // measured: structured tool calls end to end with Ollama + OpenCode
)

// Model is one catalog entry. WeightsGB is the full tag size from the Ollama registry manifest
// (including any vision projector, which Ollama loads); KV figures come from the model's config.json,
// counting only full-attention layers for hybrid models. Sources and dates: docs/MODEL_CATALOG.md.
type Model struct {
	Tag             string
	ParamsB         float64 // total parameters, billions
	ActiveParamsB   float64 // parameters used per token; equals ParamsB for dense models
	WeightsGB       float64 // download size of Tag in decimal GB
	KVBytesPerToken int     // f16 KV cache per context token: 2 x full-attention layers x kv_heads x head_dim x 2 bytes
	License         string  // all current entries are open-weight; docs/MODEL_CATALOG.md tracks open-source status
	ToolCalls       ToolCalls
	Evidence        string    // why it is ranked here: [M] measured by Odin, [V] vendor-reported, [T] third-party
	Sampling        *Sampling // vendor-recommended sampling for coding; nil keeps Ollama's settings
}

// Sampling is a vendor's recommended sampling for agentic coding. Ollama's OpenAI-compatible endpoint
// resets temperature and top_p to 1.0 unless the request sets them and ignores top_k, so Odin bakes all
// four into the derived context model and has OpenCode send temperature and top_p as well.
type Sampling struct {
	Temperature     float64
	TopP            float64
	TopK            int
	PresencePenalty float64
	NoThinking      bool // turn the model's thinking off through OpenCode (reasoningEffort "none")
}

// OllamaParameters are the Modelfile parameters for a derived model; a nil Sampling has none.
func (s *Sampling) OllamaParameters() map[string]any {
	if s == nil {
		return nil
	}
	return map[string]any{"temperature": s.Temperature, "top_p": s.TopP, "top_k": s.TopK, "presence_penalty": s.PresencePenalty}
}

// SamplingFor returns the catalog's recommended sampling for tag, or nil.
func SamplingFor(tag string) *Sampling {
	for _, m := range Catalog {
		if m.Tag == tag {
			return m.Sampling
		}
	}
	return nil
}

// DefaultAgentModel is used when no model has been configured. It must stay a verified, small entry.
const DefaultAgentModel = "qwen3.5:9b-q4_K_M"

// Catalog is ordered best-first: Recommend takes the first suitable entry that fits the machine.
// Pin explicit quant tags: some default tags bundle extra layers (e.g. qwen3.6:27b adds MTP + vision).
var Catalog = []Model{
	{
		Tag: "qwen3.6:27b-q4_K_M", ParamsB: 27, ActiveParamsB: 27, WeightsGB: 17.42, KVBytesPerToken: 65536,
		License: "Apache-2.0", ToolCalls: ToolCallsReported,
		Evidence: "Best agentic coding evidence that fits a 24 GB card: SWE-bench Verified 77.2 [V], Terminal-Bench 2.1 60.7 [T, Artificial Analysis].",
	},
	{
		Tag: "qwen3.6:35b-a3b-q4_K_M", ParamsB: 35, ActiveParamsB: 3, WeightsGB: 23.94, KVBytesPerToken: 20480,
		License: "Apache-2.0", ToolCalls: ToolCallsReported,
		Evidence: "3B-active MoE, fast where memory bandwidth is the limit: SWE-bench Verified 73.4 [V], Terminal-Bench 2.1 44.9 [T, Artificial Analysis].",
	},
	{
		Tag: "qwen3.5:35b-a3b-q4_K_M", ParamsB: 35, ActiveParamsB: 3, WeightsGB: 22.07, KVBytesPerToken: 20480,
		License: "Apache-2.0", ToolCalls: ToolCallsReported,
		Evidence: "3B-active MoE that fits 32 GB machines without a GPU: SWE-bench Verified 69.2 [V], Terminal-Bench 2.1 40.8 [T, Artificial Analysis].",
	},
	{
		Tag: "gpt-oss:20b", ParamsB: 20.9, ActiveParamsB: 3.6, WeightsGB: 13.79, KVBytesPerToken: 24576,
		License: "Apache-2.0", ToolCalls: ToolCallsBroken,
		Evidence: "Fast MoE for chat (LiveCodeBench 77.7 [T, Artificial Analysis]); not for agents: Ollama rejects tool calls it generates (ollama#17638, open).",
	},
	{
		Tag: "qwen3.5:9b-q4_K_M", ParamsB: 9, ActiveParamsB: 9, WeightsGB: 6.55, KVBytesPerToken: 32768,
		License: "Apache-2.0", ToolCalls: ToolCallsVerified,
		// Qwen's "precise coding" sampling from the Qwen3.5 model card [V] (Ollama ships temperature 1.0 and
		// presence_penalty 1.5), with thinking off: Odin's benchmark scored it at least as high and faster [M].
		Sampling: &Sampling{Temperature: 0.6, TopP: 0.95, TopK: 20, NoThinking: true},
		Evidence: "Measured by Odin [M]: passed 17 and 15 of 24 held-out Exercism Go runs through Baldur in two rounds (median under 2 min) and the Stage 0 task, 100% GPU at 64k on 12 GB; Terminal-Bench 2.1 29.2 [T, Artificial Analysis].",
	},
	{
		Tag: "qwen3.5:4b-q4_K_M", ParamsB: 4, ActiveParamsB: 4, WeightsGB: 3.32, KVBytesPerToken: 32768,
		License: "Apache-2.0", ToolCalls: ToolCallsReported,
		Evidence: "Smallest model with credible agentic evidence: BFCL-V4 50.3 [V], Terminal-Bench 2.1 25.8 [T, Artificial Analysis].",
	},
}
