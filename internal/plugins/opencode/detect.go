package opencode

import (
	"github.com/danielsampar12/odin/internal/plugins"
	"github.com/danielsampar12/odin/internal/system"
)

func Detect() plugins.Status {
	status := system.DetectCommand("opencode")
	return plugins.Status{
		Name:      "OpenCode",
		Command:   "opencode",
		Installed: status.Installed,
		Path:      status.Path,
	}
}
