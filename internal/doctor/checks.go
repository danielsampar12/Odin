package doctor

import (
	"context"
	"os"
	"strings"

	"github.com/danielsampar12/odin/internal/config"
	"github.com/danielsampar12/odin/internal/plugins"
	"github.com/danielsampar12/odin/internal/plugins/mempalace"
	"github.com/danielsampar12/odin/internal/plugins/ollama"
	"github.com/danielsampar12/odin/internal/plugins/opencode"
	shellplugin "github.com/danielsampar12/odin/internal/plugins/shell"
	"github.com/danielsampar12/odin/internal/system"
)

func Collect(ctx context.Context, cwd string) (Result, error) {
	osInfo := system.DetectOS()
	shellInfo := system.DetectShell()
	ramGB := system.DetectRAMGB(ctx)
	gpu := system.DetectGPU(ctx)
	git := statusFromCommand("Git", "git")
	starshipPrompt := shellplugin.Detect(osInfo.HomeDir)
	inGitRepo, gitRoot := detectGitRepo(ctx, cwd, git.Installed)
	globalConfigPath := config.GlobalConfigPath()
	projectConfigPath := config.ProjectConfigPath(cwd)
	ollamaBaseURL := config.ResolveGlobalRuntimeBaseURL(globalConfigPath, ollama.DefaultBaseURL)
	ollamaAPI := ollama.Probe(ctx, ollamaBaseURL)

	return Result{
		CurrentDir: cwd,
		OS:         osInfo,
		Shell:      shellInfo,
		RAMGB:      ramGB,
		GPU:        gpu,
		InGitRepo:  inGitRepo,
		GitRoot:    gitRoot,
		Tools: map[string]plugins.Status{
			"git":        git,
			"ollama":     ollama.Detect(),
			"opencode":   opencode.Detect(),
			"mempalace":  mempalace.Detect(),
			"starship":   starshipPrompt.Starship,
			"nvidia-smi": gpu.CommandStatus(),
		},
		Powerlevel10kConfigured: starshipPrompt.Powerlevel10kConfigured,
		Powerlevel10kSource:     starshipPrompt.Powerlevel10kSource,
		Ollama:                  ollamaAPI,
		GlobalConfig: FileStatus{
			Path:   globalConfigPath,
			Exists: fileExists(globalConfigPath),
		},
		ProjectConfig: FileStatus{
			Path:   projectConfigPath,
			Exists: fileExists(projectConfigPath),
		},
	}, nil
}

func detectGitRepo(ctx context.Context, cwd string, gitInstalled bool) (bool, string) {
	if !gitInstalled {
		return false, ""
	}

	output, err := system.RunCommand(ctx, cwd, "git", "rev-parse", "--show-toplevel")
	if err != nil {
		return false, ""
	}

	root := strings.TrimSpace(output)
	if root == "" {
		return false, ""
	}

	return true, root
}

func statusFromCommand(name, command string) plugins.Status {
	status := system.DetectCommand(command)
	return plugins.Status{
		Name:      name,
		Command:   command,
		Installed: status.Installed,
		Path:      status.Path,
	}
}

func fileExists(path string) bool {
	if _, err := os.Stat(path); err == nil {
		return true
	}

	return false
}
