package cli

import (
	"fmt"

	"github.com/danielsampar12/odin/internal/doctor"
	"github.com/spf13/cobra"
)

func newDoctorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Diagnose the current system and project",
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := workingDir()
			if err != nil {
				return err
			}

			result, err := doctor.Collect(cmd.Context(), cwd)
			if err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			fmt.Fprintln(out, "Odin doctor")
			fmt.Fprintln(out)
			fmt.Fprintln(out, "System")
			fmt.Fprintf(out, "- OS: %s\n", result.OS.Name)
			fmt.Fprintf(out, "- Architecture: %s\n", result.OS.Arch)
			fmt.Fprintf(out, "- Shell: %s\n", shellDisplay(result.Shell.Path, result.Shell.Name))
			if result.RAMGB > 0 {
				fmt.Fprintf(out, "- RAM: %dGB\n", result.RAMGB)
			} else {
				fmt.Fprintln(out, "- RAM: unavailable")
			}
			fmt.Fprintf(out, "- GPU: %s\n", result.GPU.Summary)
			fmt.Fprintln(out)
			fmt.Fprintln(out, "Project")
			fmt.Fprintf(out, "- Current directory: %s\n", result.CurrentDir)
			fmt.Fprintf(out, "- Inside Git repository: %s\n", yesNo(result.InGitRepo))
			if result.GitRoot != "" {
				fmt.Fprintf(out, "- Git root: %s\n", displayPath(result.GitRoot))
			}
			fmt.Fprintf(out, "- Global config: %s\n", fileStatusLabel(result.GlobalConfig.Path, result.GlobalConfig.Exists))
			fmt.Fprintf(out, "- Project config: %s\n", fileStatusLabel(result.ProjectConfig.Path, result.ProjectConfig.Exists))
			fmt.Fprintln(out)
			fmt.Fprintln(out, "Tooling")

			for _, name := range []string{"git", "ollama", "opencode", "mempalace", "starship", "nvidia-smi"} {
				tool := result.Tool(name)
				if tool.Path != "" {
					fmt.Fprintf(out, "- %s: %s (%s)\n", tool.Name, installedLabel(tool.Installed), tool.Path)
					continue
				}
				fmt.Fprintf(out, "- %s: %s\n", tool.Name, installedLabel(tool.Installed))
			}

			if result.Powerlevel10kConfigured {
				fmt.Fprintf(out, "- Powerlevel10k: configured (%s)\n", displayPath(result.Powerlevel10kSource))
			} else {
				fmt.Fprintln(out, "- Powerlevel10k: not detected")
			}
			fmt.Fprintf(out, "- Ollama API: %s\n", ollamaAPIStatusLine(result.Ollama))
			if result.Ollama.APIAvailable {
				fmt.Fprintf(out, "- Ollama models: %s\n", summarizeModelNames(result.Ollama.Models, 5))
			}

			return nil
		},
	}
}
