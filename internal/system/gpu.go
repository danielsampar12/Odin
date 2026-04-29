package system

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/danielsampar12/odin/internal/plugins"
)

type GPUInfo struct {
	CommandInstalled bool
	CommandPath      string
	Detected         bool
	VRAMGB           int
	Summary          string
}

func DetectGPU(ctx context.Context) GPUInfo {
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
