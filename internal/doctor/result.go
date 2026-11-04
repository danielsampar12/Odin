package doctor

import (
	"github.com/danielsampar12/odin/internal/plugins"
	ollamaplugin "github.com/danielsampar12/odin/internal/plugins/ollama"
	"github.com/danielsampar12/odin/internal/system"
)

type FileStatus struct {
	Error  string // set when the file exists but cannot be parsed
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
	OpenCodeGeneratedConfig FileStatus
	LegacyShellFunction     string // rc file where Odin v1 defined an odin() shell function, if any
}

func (r Result) Tool(name string) plugins.Status {
	if tool, ok := r.Tools[name]; ok {
		return tool
	}

	return plugins.Status{Name: name, Command: name}
}
