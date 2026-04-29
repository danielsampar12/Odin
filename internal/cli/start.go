package cli

import (
	"fmt"

	"github.com/danielsampar12/odin/internal/doctor"
	opencodeplugin "github.com/danielsampar12/odin/internal/plugins/opencode"
	"github.com/spf13/cobra"
)

func newStartCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "start",
		Short: "Scaffold Odin stack startup for the current project",
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
			fmt.Fprintln(out, "Odin start is currently scaffolded.")
			fmt.Fprintln(out)
			fmt.Fprintln(out, "Planned actions:")
			fmt.Fprintf(out, "- Load global config: %s\n", fileStatusLabel(result.GlobalConfig.Path, result.GlobalConfig.Exists))
			fmt.Fprintf(out, "- Load project config: %s\n", fileStatusLabel(result.ProjectConfig.Path, result.ProjectConfig.Exists))
			fmt.Fprintf(out, "- Check Ollama API: %s\n", ollamaAPIStatusLine(result.Ollama))
			fmt.Fprintf(out, "- Check MemPalace: %s\n", installedLabel(result.Tool("mempalace").Installed))
			fmt.Fprintf(out, "- Check OpenCode: %s\n", installedLabel(result.Tool("opencode").Installed))
			fmt.Fprintf(out, "- Check generated OpenCode config: %s\n", fileStatusLabel(result.OpenCodeGeneratedConfig.Path, result.OpenCodeGeneratedConfig.Exists))
			fmt.Fprintf(out, "- Planned launch command: %s\n", opencodeplugin.LaunchCommand())
			fmt.Fprintln(out)

			if !result.GlobalConfig.Exists {
				fmt.Fprintln(out, "Global config is missing. Run `odin setup` first.")
			}
			if !result.ProjectConfig.Exists {
				fmt.Fprintln(out, "Project config is missing. Run `odin init` in this repository.")
			}
			if !result.Ollama.APIAvailable {
				fmt.Fprintln(out, "Ollama API is not responding. Start Ollama before expecting Odin to launch OpenCode.")
			}
			if !result.Tool("mempalace").Installed {
				fmt.Fprintln(out, "MemPalace is not installed. Odin recommends it as the primary memory provider.")
			}
			if !result.Tool("opencode").Installed {
				fmt.Fprintln(out, "OpenCode is not installed. Install it before expecting Odin to launch the coding agent.")
			} else {
				fmt.Fprintf(out, "OpenCode detected at %s. Automatic launch is intentionally not implemented yet.\n", result.Tool("opencode").Path)
			}
			if !result.OpenCodeGeneratedConfig.Exists {
				fmt.Fprintln(out, "Generated OpenCode config is missing. Run `odin opencode generate` first.")
			}

			return nil
		},
	}
}
