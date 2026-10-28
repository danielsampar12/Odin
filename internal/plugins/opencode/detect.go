package opencode

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/danielsampar12/odin/internal/plugins"
	"github.com/danielsampar12/odin/internal/system"
)

func Detect() plugins.Status {
	status := system.DetectCommand("opencode")
	result := plugins.Status{
		Name:      "OpenCode",
		Command:   "opencode",
		Installed: status.Installed,
		Path:      status.Path,
	}

	if !status.Installed {
		return result
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	output, err := system.RunCommand(ctx, "", "opencode", "--version")
	if err != nil {
		message := strings.TrimSpace(output)
		if message == "" {
			message = err.Error()
		}
		result.Details = fmt.Sprintf("version check failed: %s", message)
		return result
	}

	result.Details = strings.TrimSpace(output)
	return result
}

func Working(status plugins.Status) bool {
	return status.Installed && !strings.HasPrefix(status.Details, "version check failed:")
}
