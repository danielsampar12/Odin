package mempalace

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/danielsampar12/odin/internal/plugins"
	"github.com/danielsampar12/odin/internal/system"
)

const helperTimeout = 2 * time.Second

func Detect() plugins.Status {
	status := system.DetectCommand("mempalace")
	result := plugins.Status{
		Name:      "MemPalace",
		Command:   "mempalace",
		Installed: status.Installed,
		Path:      status.Path,
	}

	if !status.Installed {
		return result
	}

	ctx, cancel := context.WithTimeout(context.Background(), helperTimeout)
	defer cancel()

	output, err := system.RunCommand(ctx, "", "mempalace", "mcp")
	if err != nil {
		message := strings.TrimSpace(output)
		if message == "" {
			message = err.Error()
		}
		result.Details = fmt.Sprintf("mcp helper check failed: %s", message)
		return result
	}

	result.Details = summarizeMCPHelper(output)
	return result
}

func Working(status plugins.Status) bool {
	return status.Installed && !strings.HasPrefix(status.Details, "mcp helper check failed:")
}

func summarizeMCPHelper(output string) string {
	firstLine := strings.TrimSpace(strings.Split(output, "\n")[0])
	if firstLine == "" {
		return "mcp helper available"
	}

	if len(firstLine) > 72 {
		firstLine = firstLine[:69] + "..."
	}

	return "mcp helper available: " + firstLine
}
