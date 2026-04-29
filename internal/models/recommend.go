package models

import (
	"fmt"

	"github.com/danielsampar12/odin/internal/system"
)

type Recommendation struct {
	Recommended    string
	Fallback       string
	OptionalLarger string
	Reason         string
}

// RecommendCodingModel intentionally uses conservative heuristics.
// The goal is a safe default for local development, not the largest model
// that might barely fit on a given machine.
func RecommendCodingModel(ramGB int, gpu system.GPUInfo) Recommendation {
	if gpu.Detected {
		switch {
		case gpu.VRAMGB >= 20:
			return Recommendation{
				Recommended: "qwen3-coder:30b",
				Fallback:    "qwen2.5-coder:14b-instruct-q5_K_M",
				Reason:      fmt.Sprintf("Detected an NVIDIA GPU with about %dGB VRAM, which is enough to start with Odin's larger coding-model tier.", gpu.VRAMGB),
			}
		case gpu.VRAMGB >= 12:
			return Recommendation{
				Recommended:    "qwen2.5-coder:14b-instruct-q5_K_M",
				Fallback:       "qwen2.5-coder:7b",
				OptionalLarger: "qwen3-coder:30b",
				Reason:         fmt.Sprintf("Detected an NVIDIA GPU with about %dGB VRAM, which is a good fit for the 14B Qwen coder tier.", gpu.VRAMGB),
			}
		case gpu.VRAMGB >= 8:
			return Recommendation{
				Recommended: "qwen2.5-coder:7b",
				Fallback:    "qwen2.5-coder:3b",
				Reason:      fmt.Sprintf("Detected an NVIDIA GPU with about %dGB VRAM, so the 7B tier is the safest coding default.", gpu.VRAMGB),
			}
		default:
			return Recommendation{
				Recommended: "qwen2.5-coder:3b",
				Fallback:    "qwen2.5-coder:3b",
				Reason:      fmt.Sprintf("Detected a small NVIDIA GPU with about %dGB VRAM, so Odin is staying with the smallest safe local coding tier.", gpu.VRAMGB),
			}
		}
	}

	switch {
	case ramGB >= 32:
		return Recommendation{
			Recommended:    "qwen2.5-coder:7b",
			Fallback:       "qwen2.5-coder:3b",
			OptionalLarger: "qwen2.5-coder:14b-instruct-q5_K_M",
			Reason:         fmt.Sprintf("No dedicated GPU was detected, but %dGB RAM is enough to start with a 7B CPU-friendly coding model.", ramGB),
		}
	case ramGB >= 16:
		return Recommendation{
			Recommended: "qwen2.5-coder:7b",
			Fallback:    "qwen2.5-coder:3b",
			Reason:      fmt.Sprintf("No dedicated GPU was detected, but %dGB RAM can still handle a 7B model if you accept slower CPU inference.", ramGB),
		}
	default:
		return Recommendation{
			Recommended: "qwen2.5-coder:3b",
			Fallback:    "qwen2.5-coder:3b",
			Reason:      fmt.Sprintf("No dedicated GPU was detected and only about %dGB RAM is available, so Odin is recommending the lightest coding model tier.", ramGB),
		}
	}
}
