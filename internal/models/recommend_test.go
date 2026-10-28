package models

import (
	"strings"
	"testing"

	"github.com/danielsampar12/odin/internal/system"
)

// Hardware profiles from the 2026-10-28 model research; expectations follow its per-profile table
// where the catalog has the same entries.
func TestRecommendScenarios(t *testing.T) {
	apple := system.GPUInfo{AppleSilicon: true}
	nvidia := func(vram int) system.GPUInfo { return system.GPUInfo{Detected: true, VRAMGB: vram} }

	testCases := []struct {
		name         string
		ramGB        int
		gpu          system.GPUInfo
		useCase      UseCase
		wantModel    string
		wantFallback string
		wantOptional string
		wantContext  int
		wantReason   string
	}{
		{"cpu only 8gb", 8, system.GPUInfo{}, UseCaseAgent, "qwen3.5:4b-q4_K_M", "qwen3.5:4b-q4_K_M", "", 32768, "minutes per agent turn"},
		{"cpu only 6gb keeps the minimum context", 6, system.GPUInfo{}, UseCaseAgent, "qwen3.5:4b-q4_K_M", "qwen3.5:4b-q4_K_M", "", 16384, "fits with 16384 tokens"},
		{"cpu only 16gb", 16, system.GPUInfo{}, UseCaseAgent, "qwen3.5:4b-q4_K_M", "qwen3.5:4b-q4_K_M", "", 32768, "No dedicated GPU"},
		{"cpu only 32gb picks 3b-active moe", 32, system.GPUInfo{}, UseCaseAgent, "qwen3.5:35b-a3b-q4_K_M", "qwen3.5:4b-q4_K_M", "", 32768, "no verified model is suitable"},
		{"nvidia 8gb", 32, nvidia(8), UseCaseAgent, "qwen3.5:4b-q4_K_M", "qwen3.5:4b-q4_K_M", "qwen3.5:9b-q4_K_M", 32768, "not verified"},
		{"nvidia 12gb", 32, nvidia(12), UseCaseAgent, "qwen3.5:9b-q4_K_M", "qwen3.5:4b-q4_K_M", "", 32768, "Measured by Odin"},
		{"nvidia 16gb", 32, nvidia(16), UseCaseAgent, "qwen3.5:9b-q4_K_M", "qwen3.5:4b-q4_K_M", "", 32768, "NVIDIA GPU with about 16GB"},
		{"nvidia 24gb", 64, nvidia(24), UseCaseAgent, "qwen3.6:27b-q4_K_M", "qwen3.5:9b-q4_K_M", "", 32768, "measured fallback"},
		{"apple 16gb", 16, apple, UseCaseAgent, "qwen3.5:9b-q4_K_M", "qwen3.5:4b-q4_K_M", "", 32768, "Apple Silicon"},
		{"apple 32gb", 32, apple, UseCaseAgent, "qwen3.6:27b-q4_K_M", "qwen3.5:9b-q4_K_M", "", 32768, "unified memory"},
		{"apple 48gb", 48, apple, UseCaseAgent, "qwen3.6:27b-q4_K_M", "qwen3.5:9b-q4_K_M", "", 32768, "measured fallback"},
		{"chat nvidia 16gb allows a tool-broken model", 32, nvidia(16), UseCaseChat, "gpt-oss:20b", "qwen3.5:9b-q4_K_M", "", 8192, "fits with 8192 tokens"},
		{"chat cpu only 32gb", 32, system.GPUInfo{}, UseCaseChat, "qwen3.5:35b-a3b-q4_K_M", "gpt-oss:20b", "", 8192, "No dedicated GPU"},
		{"nothing fits", 4, system.GPUInfo{}, UseCaseAgent, "qwen3.5:4b-q4_K_M", "qwen3.5:4b-q4_K_M", "", 16384, "No catalog model fits"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := Recommend(tc.ramGB, tc.gpu, tc.useCase)
			if got.Recommended != tc.wantModel || got.Fallback != tc.wantFallback || got.OptionalLarger != tc.wantOptional || got.Context != tc.wantContext {
				t.Fatalf("got %s / fallback %s / optional %q / ctx %d, want %s / %s / %q / %d",
					got.Recommended, got.Fallback, got.OptionalLarger, got.Context,
					tc.wantModel, tc.wantFallback, tc.wantOptional, tc.wantContext)
			}
			if !strings.Contains(got.Reason, tc.wantReason) {
				t.Fatalf("reason = %q, want substring %q", got.Reason, tc.wantReason)
			}
		})
	}
}

func TestRecommendCodingModelIsAgentUseCase(t *testing.T) {
	if got, want := RecommendCodingModel(32, gpu12GB), Recommend(32, gpu12GB, UseCaseAgent); got != want {
		t.Fatalf("RecommendCodingModel = %+v, want the agent recommendation %+v", got, want)
	}
}

func TestCatalogInvariants(t *testing.T) {
	seen := map[string]bool{}
	for _, m := range Catalog {
		if seen[m.Tag] {
			t.Fatalf("duplicate catalog tag %s", m.Tag)
		}
		seen[m.Tag] = true
		if m.WeightsGB <= 0 || m.KVBytesPerToken <= 0 || m.ActiveParamsB <= 0 || m.ActiveParamsB > m.ParamsB || m.License == "" || m.Evidence == "" {
			t.Fatalf("incomplete catalog entry %+v", m)
		}
	}

	var defaultModel Model
	for _, m := range Catalog {
		if m.Tag == DefaultAgentModel {
			defaultModel = m
		}
	}
	if defaultModel.ToolCalls != ToolCallsVerified {
		t.Fatalf("DefaultAgentModel %s must be a catalog entry with verified tool calls", DefaultAgentModel)
	}

	for _, useCase := range []UseCase{UseCaseAgent, UseCaseChat} {
		for _, cpuOnly := range []bool{false, true} {
			found := false
			for _, m := range Catalog {
				found = found || suitable(m, useCase, cpuOnly)
			}
			if !found {
				t.Fatalf("no suitable catalog entry for use case %s (cpu only %v)", useCase, cpuOnly)
			}
		}
	}
}
