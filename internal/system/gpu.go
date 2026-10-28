package system

import (
	"context"
	"fmt"
	"runtime"
	"strconv"
	"strings"

	"github.com/danielsampar12/odin/internal/plugins"
)

type GPUInfo struct {
	CommandInstalled bool
	CommandPath      string
	Detected         bool // a discrete GPU (NVIDIA) with a known VRAM figure
	AppleSilicon     bool // Apple Silicon GPU: Metal-accelerated, unified memory
	VRAMGB           int
	Summary          string
}

func DetectGPU(ctx context.Context) GPUInfo {
	// Apple Silicon has an integrated, Metal-accelerated GPU with unified memory.
	// There's no nvidia-smi and no discrete VRAM figure, so we detect it
	// separately and let the recommender size models from unified (system) memory.
	if isAppleSilicon(ctx) {
		return GPUInfo{
			AppleSilicon: true,
			Summary:      "Apple Silicon GPU detected (Metal, unified memory)",
		}
	}

	command := DetectCommand("nvidia-smi")
	info := GPUInfo{
		CommandInstalled: command.Installed,
		CommandPath:      command.Path,
		Summary:          "No dedicated GPU detected",
	}

	if !command.Installed {
		return info
	}

	output, err := RunCommand(ctx, "", "nvidia-smi", "--query-gpu=memory.total", "--format=csv,noheader,nounits")
	if err != nil {
		info.Summary = "nvidia-smi installed, but GPU details are unavailable"
		return info
	}

	maxVRAMMB := 0
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		value, err := strconv.Atoi(line)
		if err != nil {
			continue
		}

		if value > maxVRAMMB {
			maxVRAMMB = value
		}
	}

	if maxVRAMMB == 0 {
		info.Summary = "nvidia-smi installed, but no GPU details were returned"
		return info
	}

	info.Detected = true
	info.VRAMGB = (maxVRAMMB + 1023) / 1024
	info.Summary = fmt.Sprintf("NVIDIA GPU detected (%dGB VRAM)", info.VRAMGB)
	return info
}

func (g GPUInfo) CommandStatus() plugins.Status {
	return plugins.Status{
		Name:      "nvidia-smi",
		Command:   "nvidia-smi",
		Installed: g.CommandInstalled,
		Path:      g.CommandPath,
	}
}

// isAppleSilicon reports whether the host is an Apple Silicon Mac. We can't rely
// on runtime.GOARCH: an amd64 Go toolchain running under Rosetta reports "amd64"
// even on Apple Silicon. hw.optional.arm64 reflects the hardware, not the process.
func isAppleSilicon(ctx context.Context) bool {
	if runtime.GOOS != "darwin" {
		return false
	}

	output, err := RunCommand(ctx, "", "sysctl", "-n", "hw.optional.arm64")
	if err != nil {
		return false
	}

	return strings.TrimSpace(output) == "1"
}
