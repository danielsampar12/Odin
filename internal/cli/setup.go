package cli

import (
	"fmt"

	"github.com/danielsampar12/odin/internal/companions"
	"github.com/danielsampar12/odin/internal/config"
	"github.com/danielsampar12/odin/internal/doctor"
	"github.com/spf13/cobra"
)

type setupRecommendation struct {
	Profile        string
	Model          string
	FallbackModel  string
	Agent          string
	Memory         string
	CompanionKey   string
	CompanionName  string
	ShellProvider  string
	ShellEnabled   bool
	RuntimeBaseURL string
}

func newSetupCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "setup",
		Short: "Scaffold global Odin setup",
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := workingDir()
			if err != nil {
				return err
			}

			result, err := doctor.Collect(cmd.Context(), cwd)
			if err != nil {
				return err
			}

			recommendation := buildSetupRecommendation(result)
			globalDir := config.GlobalDir()
			globalConfigPath := config.GlobalConfigPath()

			dirCreated, err := ensureDir(globalDir)
			if err != nil {
				return err
			}

			globalConfigCreated, err := writeFileIfMissing(globalConfigPath, config.DefaultGlobalConfig(config.GlobalSettings{
				Profile:          recommendation.Profile,
				RuntimeProvider:  "ollama",
				RuntimeBaseURL:   recommendation.RuntimeBaseURL,
				AgentProvider:    "opencode",
				MemoryProvider:   "mempalace",
				ShellProvider:    recommendation.ShellProvider,
				ShellEnabled:     recommendation.ShellEnabled,
				ModelDefault:     recommendation.Model,
				ModelFallback:    recommendation.FallbackModel,
				CompanionDefault: recommendation.CompanionKey,
			}), 0o644)
			if err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			fmt.Fprintln(out, "Odin setup is currently scaffolded. It does not install external tools yet.")
			fmt.Fprintln(out)
			fmt.Fprintln(out, "Odin checked your machine.")
			fmt.Fprintln(out)
			fmt.Fprintln(out, "You have:")
			if result.RAMGB > 0 {
				fmt.Fprintf(out, "- %dGB RAM\n", result.RAMGB)
			} else {
				fmt.Fprintln(out, "- RAM information unavailable")
			}
			fmt.Fprintf(out, "- %s\n", result.GPU.Summary)
			fmt.Fprintf(out, "- Ollama %s\n", installedLabel(result.Tool("ollama").Installed))
			fmt.Fprintf(out, "- OpenCode %s\n", installedLabel(result.Tool("opencode").Installed))
			fmt.Fprintf(out, "- MemPalace %s\n", installedLabel(result.Tool("mempalace").Installed))
			fmt.Fprintln(out)
			fmt.Fprintln(out, "Recommended setup:")
			fmt.Fprintf(out, "- Profile: %s\n", recommendation.Profile)
			fmt.Fprintf(out, "- Model: %s\n", recommendation.Model)
			fmt.Fprintf(out, "- Agent: %s\n", recommendation.Agent)
			fmt.Fprintf(out, "- Memory: %s\n", recommendation.Memory)
			fmt.Fprintf(out, "- Companion: %s\n", recommendation.CompanionName)
			fmt.Fprintln(out)
			fmt.Fprintln(out, "Global config:")
			if dirCreated {
				fmt.Fprintf(out, "- Created %s\n", displayPath(globalDir))
			} else {
				fmt.Fprintf(out, "- %s already exists\n", displayPath(globalDir))
			}
			if globalConfigCreated {
				fmt.Fprintf(out, "- Created %s\n", displayPath(globalConfigPath))
			} else {
				fmt.Fprintf(out, "- Kept existing %s\n", displayPath(globalConfigPath))
			}
			fmt.Fprintln(out)
			fmt.Fprintln(out, "Install? [Y/n]")
			fmt.Fprintln(out, "Not implemented yet. This scaffold only reports the planned global setup flow.")

			if result.InGitRepo {
				fmt.Fprintln(out)
				fmt.Fprintln(out, "You are inside a Git repository. Initialize Odin for this project too? [Y/n]")
				fmt.Fprintln(out, "Future behavior: `odin setup` will be able to reuse the same scaffold path as `odin init`.")
			}

			return nil
		},
	}
}

func buildSetupRecommendation(result doctor.Result) setupRecommendation {
	profile := detectProfile(result.CurrentDir, result.InGitRepo)
	companion := companions.DefaultForProfile(profile)
	model, fallback := recommendModels(result)
	shellProvider, shellEnabled := recommendShellProvider(result)

	return setupRecommendation{
		Profile:        profile,
		Model:          model,
		FallbackModel:  fallback,
		Agent:          "OpenCode",
		Memory:         "MemPalace",
		CompanionKey:   companion.Key,
		CompanionName:  companion.Name,
		ShellProvider:  shellProvider,
		ShellEnabled:   shellEnabled,
		RuntimeBaseURL: "http://localhost:11434",
	}
}

func recommendModels(result doctor.Result) (string, string) {
	if result.GPU.Detected {
		switch {
		case result.GPU.VRAMGB >= 20:
			return "qwen3-coder:30b", "qwen2.5-coder:14b-instruct-q5_K_M"
		case result.GPU.VRAMGB >= 12:
			return "qwen2.5-coder:14b-instruct-q5_K_M", "qwen2.5-coder:7b"
		case result.GPU.VRAMGB >= 8:
			return "qwen2.5-coder:7b", "qwen2.5-coder:3b"
		default:
			return "qwen2.5-coder:3b", "qwen2.5-coder:3b"
		}
	}

	switch {
	case result.RAMGB >= 32:
		return "qwen2.5-coder:14b-instruct-q5_K_M", "qwen2.5-coder:7b"
	case result.RAMGB >= 16:
		return "qwen2.5-coder:7b", "qwen2.5-coder:3b"
	default:
		return "qwen2.5-coder:3b", "qwen2.5-coder:3b"
	}
}

func recommendShellProvider(result doctor.Result) (string, bool) {
	if result.Powerlevel10kConfigured {
		return "p10k", true
	}
	if result.Tool("starship").Installed {
		return "starship", true
	}
	return "p10k", false
}
