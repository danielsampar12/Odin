package system

import (
	"context"
	"os"
	"runtime"
	"strconv"
	"strings"
)

type OSInfo struct {
	Name    string
	Arch    string
	HomeDir string
}

func DetectOS() OSInfo {
	home, _ := os.UserHomeDir()
	return OSInfo{
		Name:    runtime.GOOS,
		Arch:    runtime.GOARCH,
		HomeDir: home,
	}
}

func DetectRAMGB(ctx context.Context) int {
	switch runtime.GOOS {
	case "linux":
		body, err := os.ReadFile("/proc/meminfo")
		if err != nil {
			return 0
		}

		for _, line := range strings.Split(string(body), "\n") {
			if !strings.HasPrefix(line, "MemTotal:") {
				continue
			}

			fields := strings.Fields(line)
			if len(fields) < 2 {
				return 0
			}

			memKB, err := strconv.Atoi(fields[1])
			if err != nil {
				return 0
			}

			const kbPerGB = 1024 * 1024
			return (memKB + kbPerGB - 1) / kbPerGB
		}
	case "darwin":
		output, err := RunCommand(ctx, "", "sysctl", "-n", "hw.memsize")
		if err != nil {
			return 0
		}

		memBytes, err := strconv.ParseInt(strings.TrimSpace(output), 10, 64)
		if err != nil {
			return 0
		}

		const bytesPerGB = 1024 * 1024 * 1024
		return int((memBytes + bytesPerGB - 1) / bytesPerGB)
	}

	return 0
}
