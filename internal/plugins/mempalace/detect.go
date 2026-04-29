package mempalace

import (
	"github.com/danielsampar12/odin/internal/plugins"
	"github.com/danielsampar12/odin/internal/system"
)

func Detect() plugins.Status {
	status := system.DetectCommand("mempalace")
	return plugins.Status{
		Name:      "MemPalace",
		Command:   "mempalace",
		Installed: status.Installed,
		Path:      status.Path,
	}
}
