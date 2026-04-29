package ollama

import (
	"github.com/danielsampar12/odin/internal/plugins"
	"github.com/danielsampar12/odin/internal/system"
)

func Detect() plugins.Status {
	status := system.DetectCommand("ollama")
	return plugins.Status{
		Name:      "Ollama",
		Command:   "ollama",
		Installed: status.Installed,
		Path:      status.Path,
	}
}
