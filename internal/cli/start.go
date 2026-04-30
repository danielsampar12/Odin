package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/danielsampar12/odin/internal/companions"
	"github.com/danielsampar12/odin/internal/doctor"
	ollamaplugin "github.com/danielsampar12/odin/internal/plugins/ollama"
	opencodeplugin "github.com/danielsampar12/odin/internal/plugins/opencode"
	"github.com/spf13/cobra"
)

func newStartCmd() *cobra.Command {
	var execRequested bool

	cmd := &cobra.Command{
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

			scaffold, err := opencodeplugin.ResolveProjectScaffold(cwd)
			if err != nil {
				return err
			}

			opencodeTool := result.Tool("opencode")
			out := cmd.OutOrStdout()
			fmt.Fprintln(out, "Odin start")
			fmt.Fprintln(out)
			fmt.Fprintln(out, "Planned actions:")
			fmt.Fprintf(out, "- Load global config: %s\n", fileStatusLabel(result.GlobalConfig.Path, result.GlobalConfig.Exists))
			fmt.Fprintf(out, "- Load project config: %s\n", fileStatusLabel(result.ProjectConfig.Path, result.ProjectConfig.Exists))
			fmt.Fprintf(out, "- Check Ollama API: %s\n", ollamaAPIStatusLine(result.Ollama))
			fmt.Fprintf(out, "- Check MemPalace: %s\n", installedLabel(result.Tool("mempalace").Installed))
			fmt.Fprintf(out, "- Check OpenCode: %s\n", installedLabel(opencodeTool.Installed))
			fmt.Fprintf(out, "- Check generated OpenCode config: %s\n", fileStatusLabel(result.OpenCodeGeneratedConfig.Path, result.OpenCodeGeneratedConfig.Exists))
			fmt.Fprintf(out, "- Planned launch command: %s\n", opencodeplugin.LaunchCommand(opencodeTool.Path))
			fmt.Fprintln(out)
			printStartSummary(out, cwd, scaffold, opencodeTool.Path)
			fmt.Fprintln(out)

			if !result.GlobalConfig.Exists {
				fmt.Fprintln(out, "Global config is missing. Run `odin setup` first.")
			}
			if !result.ProjectConfig.Exists {
				fmt.Fprintln(out, "Project config is missing. Run `odin init` in this repository.")
			}
			if !result.Ollama.APIAvailable {
				fmt.Fprintf(out, "Ollama API is not responding at %s.\n", result.Ollama.BaseURL)
				fmt.Fprintln(out, "Start Ollama and try again.")
			}
			if !result.Tool("mempalace").Installed {
				fmt.Fprintln(out, "MemPalace is not installed. Odin recommends it as the primary memory provider.")
			}
			if !opencodeTool.Installed {
				fmt.Fprintln(out, "OpenCode is not installed. Install it before expecting Odin to launch the coding agent.")
			} else if !opencodeplugin.Working(opencodeTool) {
				fmt.Fprintf(out, "OpenCode was found at %s, but the binary did not pass a version check.\n", opencodeTool.Path)
				if opencodeTool.Details != "" {
					fmt.Fprintf(out, "Details: %s\n", opencodeTool.Details)
				}
			} else {
				fmt.Fprintf(out, "OpenCode detected at %s.\n", opencodeTool.Path)
			}
			if !result.OpenCodeGeneratedConfig.Exists {
				fmt.Fprintln(out, "Generated OpenCode config is missing. Run `odin opencode generate` first.")
			}
			if result.Ollama.APIAvailable && !ollamaModelInstalled(result.Ollama.Models, scaffold.Model) {
				fmt.Fprintf(out, "Configured model %q is not installed locally. Pull it with `odin model pull %s`.\n", scaffold.Model, scaffold.Model)
			}

			if !execRequested {
				fmt.Fprintln(out)
				fmt.Fprintln(out, "Run `odin start --exec` to launch OpenCode with the Odin-generated project config.")
				return nil
			}

			if !result.ProjectConfig.Exists || !result.OpenCodeGeneratedConfig.Exists || !result.Ollama.APIAvailable || !opencodeTool.Installed || !opencodeplugin.Working(opencodeTool) {
				return nil
			}

			fmt.Fprintln(out)
			fmt.Fprintln(out, "Launching OpenCode now.")
			return opencodeplugin.Launch(cmd.Context(), opencodeplugin.LaunchOptions{
				BinaryPath: opencodeTool.Path,
				WorkingDir: cwd,
				ConfigPath: opencodeplugin.RelativeConfigPath(),
				Stdin:      os.Stdin,
				Stdout:     os.Stdout,
				Stderr:     os.Stderr,
			})
		},
	}

	cmd.Flags().BoolVar(&execRequested, "exec", false, "Launch OpenCode after validation")
	cmd.Flags().BoolVar(&execRequested, "run", false, "Alias for --exec")
	return cmd
}

func printStartSummary(writer io.Writer, cwd string, scaffold opencodeplugin.ProjectScaffold, binaryPath string) {
	projectName := filepath.Base(cwd)
	fmt.Fprintln(writer, "Starting Odin stack.")
	fmt.Fprintln(writer)
	fmt.Fprintf(writer, "Project: %s\n", projectName)
	fmt.Fprintf(writer, "Companion: %s\n", companions.DisplayName(scaffold.Companion))
	fmt.Fprintf(writer, "Model: %s\n", scaffold.Model)
	fmt.Fprintf(writer, "Runtime: %s\n", displayRuntimeName(scaffold.RuntimeProvider))
	fmt.Fprintf(writer, "OpenCode config: %s\n", displayPath(scaffold.ConfigPath))
	fmt.Fprintln(writer)
	fmt.Fprintln(writer, "Launching:")
	fmt.Fprintf(writer, "%s\n", opencodeplugin.LaunchCommand(binaryPath))
}

func ollamaModelInstalled(models []ollamaplugin.Model, name string) bool {
	for _, model := range models {
		if model.Name == name {
			return true
		}
	}

	return false
}

func displayRuntimeName(provider string) string {
	if provider == "" {
		return "unknown"
	}

	return strings.ToUpper(provider[:1]) + provider[1:]
}
