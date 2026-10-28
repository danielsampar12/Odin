package models

import (
	"math"
	"testing"

	"github.com/danielsampar12/odin/internal/system"
)

var gpu12GB = system.GPUInfo{Detected: true, VRAMGB: 12}

// GPU memory measured with nvidia-smi on an RTX 5070 while llama-server held each model (Stage 0,
// 2026-10-28). The estimate must stay within 0.35 GiB so fit decisions match real hardware.
func TestNeedGiBMatchesStage0Measurements(t *testing.T) {
	qwen3_8b := Model{WeightsGB: 5.2, KVBytesPerToken: 2 * 36 * 8 * 128 * 2}
	qwen3_14b := Model{WeightsGB: 9.3, KVBytesPerToken: 2 * 40 * 8 * 128 * 2}
	coder7b := Model{WeightsGB: 4.7, KVBytesPerToken: 2 * 28 * 4 * 128 * 2}

	testCases := []struct {
		name        string
		model       Model
		ctx         int
		q8KV        bool
		measuredMiB float64
	}{
		{"qwen3:8b 16k", qwen3_8b, 16384, false, 7849},
		{"qwen3:8b 32k", qwen3_8b, 32768, false, 10169},
		{"qwen2.5-coder:7b 16k", coder7b, 16384, false, 6001},
		{"qwen2.5-coder:7b 32k", coder7b, 32768, false, 6913},
		{"qwen3:14b 16k q8_0", qwen3_14b, 16384, true, 10727},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := needGiB(tc.model, tc.ctx, tc.q8KV)
			want := tc.measuredMiB / 1024
			if math.Abs(got-want) > 0.35 {
				t.Fatalf("needGiB = %.2f GiB, measured %.2f GiB", got, want)
			}
		})
	}
}

// On the measured 12 GB card, qwen3:14b at 16k spilled to the CPU with an f16 KV cache and fit fully
// with q8_0; qwen3:8b fit at 32k. The budget must reproduce those outcomes.
func TestBudgetReproducesStage0Fits(t *testing.T) {
	budget, _ := memoryBudget(32, gpu12GB)
	qwen3_14b := Model{WeightsGB: 9.3, KVBytesPerToken: 163840}
	qwen3_8b := Model{WeightsGB: 5.2, KVBytesPerToken: 147456}

	if needGiB(qwen3_14b, 16384, false) <= budget {
		t.Fatalf("qwen3:14b f16 @16k should not fit a 12 GB card")
	}
	if needGiB(qwen3_14b, 16384, true) > budget {
		t.Fatalf("qwen3:14b q8_0 @16k should fit a 12 GB card")
	}
	if needGiB(qwen3_8b, 32768, false) > budget {
		t.Fatalf("qwen3:8b f16 @32k should fit a 12 GB card")
	}
}
