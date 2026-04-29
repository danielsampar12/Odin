package models

import (
	"strings"
	"testing"

	"github.com/danielsampar12/odin/internal/system"
)

func TestRecommendCodingModel(t *testing.T) {
	testCases := []struct {
		name            string
		ramGB           int
		gpu             system.GPUInfo
		wantModel       string
		wantFallback    string
		reasonSubstring string
	}{
		{
			name:            "cpu only low ram",
			ramGB:           8,
			wantModel:       "qwen2.5-coder:3b",
			wantFallback:    "qwen2.5-coder:3b",
			reasonSubstring: "lightest coding model tier",
		},
		{
			name:            "cpu only medium ram",
			ramGB:           16,
			wantModel:       "qwen2.5-coder:7b",
			wantFallback:    "qwen2.5-coder:3b",
			reasonSubstring: "slower CPU inference",
		},
		{
			name:            "gpu around 8gb",
			ramGB:           32,
			gpu:             system.GPUInfo{Detected: true, VRAMGB: 8},
			wantModel:       "qwen2.5-coder:7b",
			wantFallback:    "qwen2.5-coder:3b",
			reasonSubstring: "7B tier",
		},
		{
			name:            "gpu around 12gb",
			ramGB:           32,
			gpu:             system.GPUInfo{Detected: true, VRAMGB: 12},
			wantModel:       "qwen2.5-coder:14b-instruct-q5_K_M",
			wantFallback:    "qwen2.5-coder:7b",
			reasonSubstring: "14B Qwen coder tier",
		},
		{
			name:            "gpu large",
			ramGB:           64,
			gpu:             system.GPUInfo{Detected: true, VRAMGB: 24},
			wantModel:       "qwen3-coder:30b",
			wantFallback:    "qwen2.5-coder:14b-instruct-q5_K_M",
			reasonSubstring: "larger coding-model tier",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := RecommendCodingModel(tc.ramGB, tc.gpu)
			if got.Recommended != tc.wantModel {
				t.Fatalf("recommended model = %q, want %q", got.Recommended, tc.wantModel)
			}
			if got.Fallback != tc.wantFallback {
				t.Fatalf("fallback model = %q, want %q", got.Fallback, tc.wantFallback)
			}
			if !strings.Contains(got.Reason, tc.reasonSubstring) {
				t.Fatalf("reason = %q, want substring %q", got.Reason, tc.reasonSubstring)
			}
		})
	}
}
