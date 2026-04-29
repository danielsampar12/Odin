package doctor

import (
	"github.com/danielsampar12/odin/internal/plugins"
	ollamaplugin "github.com/danielsampar12/odin/internal/plugins/ollama"
	"github.com/danielsampar12/odin/internal/system"
)

type FileStatus struct {
	Path   string
	Exists bool
}

type Result struct {
	CurrentDir              string
	OS                      system.OSInfo
	Shell                   system.ShellInfo
	RAMGB                   int
	GPU                     system.GPUInfo
	InGitRepo               bool
	GitRoot                 string
	Tools                   map[string]plugins.Status
	Powerlevel10kConfigured bool
	Powerlevel10kSource     string
	Ollama                  ollamaplugin.APIStatus
	GlobalConfig            FileStatus
	ProjectConfig           FileStatus
}

func (r Result) Tool(name string) plugins.Status {
	if tool, ok := r.Tools[name]; ok {
		return tool
	}

	return plugins.Status{Name: name, Command: name}
}
