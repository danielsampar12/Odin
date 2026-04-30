package cli

import (
	"fmt"
	"io"

	"github.com/danielsampar12/odin/internal/companions"
	"github.com/danielsampar12/odin/internal/doctor"
	"github.com/danielsampar12/odin/internal/plugins"
	mempalaceplugin "github.com/danielsampar12/odin/internal/plugins/mempalace"
	opencodeplugin "github.com/danielsampar12/odin/internal/plugins/opencode"
	"github.com/spf13/cobra"
)

func newOpenCodeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "opencode",
		Short: "Inspect and generate OpenCode project integration",
	}

	cmd.AddCommand(
		newOpenCodeStatusCmd(),
		newOpenCodeGenerateCmd(),
	)

	return cmd
}

func newOpenCodeStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show OpenCode integration status for the current project",
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

			memoryStatus, err := mempalaceplugin.ResolveProjectStatus(cwd)
			if err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			fmt.Fprintln(out, "OpenCode status")
			fmt.Fprintln(out)
			printOpenCodeToolStatus(out, result.Tool("opencode"))
			fmt.Fprintf(out, "Generated config: %s\n", fileStatusLabel(result.OpenCodeGeneratedConfig.Path, result.OpenCodeGeneratedConfig.Exists))
			fmt.Fprintf(out, "Runtime provider: %s\n", scaffold.RuntimeProvider)
			fmt.Fprintf(out, "Runtime base URL: %s\n", scaffold.RuntimeBaseURL)
			fmt.Fprintf(out, "OpenCode provider/model: %s\n", scaffold.ProviderModel())
			fmt.Fprintf(out, "Selected companion: %s\n", companions.DisplayName(scaffold.Companion))
			fmt.Fprintf(out, "Memory provider: %s\n", scaffold.MemoryProvider)
			fmt.Fprintf(out, "Memory hall: %s\n", scaffold.MemoryHall)
			fmt.Fprintf(out, "MemPalace MCP wiring: %s\n", memoryMCPLabel(memoryStatus.OpenCodeMCPConfigured, memoryStatus.OpenCodeMCPEnabled))
			fmt.Fprintln(out, "Project instructions: AGENTS.md and .odin/rules.md")
			if scaffold.Supported() {
				fmt.Fprintf(out, "Planned launch command: %s\n", opencodeplugin.LaunchCommand(result.Tool("opencode").Path))
			} else {
				fmt.Fprintf(out, "OpenCode generation currently supports Ollama only. Current runtime provider: %s\n", scaffold.RuntimeProvider)
			}
			if !result.OpenCodeGeneratedConfig.Exists {
				fmt.Fprintln(out, "Generated config is missing. Run `odin opencode generate`.")
			} else if !memoryStatus.OpenCodeMCPConfigured {
				fmt.Fprintln(out, "MemPalace MCP is not wired into the generated config. Re-run `odin opencode generate --with-memory` to opt in.")
			}

			return nil
		},
	}
}

func newOpenCodeGenerateCmd() *cobra.Command {
	var force bool
	var withMemory bool

	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate a project-local OpenCode config scaffold",
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := workingDir()
			if err != nil {
				return err
			}

			result, err := opencodeplugin.WriteGeneratedConfig(cwd, opencodeplugin.GenerateOptions{
				Force:      force,
				WithMemory: withMemory,
			})
			if err != nil {
				if err == opencodeplugin.ErrUnmanagedGeneratedConfig {
					fmt.Fprintf(cmd.OutOrStdout(), "Refusing to overwrite %s because it is not Odin-managed. Re-run with `--force` if you want Odin to replace it.\n", displayPath(result.Path))
					return nil
				}
				return err
			}

			out := cmd.OutOrStdout()
			scaffold, err := opencodeplugin.ResolveProjectScaffold(cwd)
			if err != nil {
				return err
			}

			if result.Written && result.Updated {
				fmt.Fprintf(out, "Updated %s\n", displayPath(result.Path))
			} else if result.Written {
				fmt.Fprintf(out, "Created %s\n", displayPath(result.Path))
			} else {
				fmt.Fprintf(out, "Kept existing %s\n", displayPath(result.Path))
			}
			fmt.Fprintf(out, "OpenCode provider/model: %s\n", scaffold.ProviderModel())
			fmt.Fprintf(out, "Selected companion: %s\n", companions.DisplayName(scaffold.Companion))
			fmt.Fprintf(out, "Memory provider: %s\n", scaffold.MemoryProvider)
			fmt.Fprintf(out, "Memory hall: %s\n", scaffold.MemoryHall)
			if withMemory {
				fmt.Fprintln(out, "MemPalace MCP: enabled in generated config")
			} else {
				fmt.Fprintln(out, "MemPalace MCP: not included by default")
			}
			fmt.Fprintln(out, "Project instructions: AGENTS.md and .odin/rules.md")
			fmt.Fprintf(out, "Use with: %s\n", opencodeplugin.LaunchCommand(""))

			return nil
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Overwrite the generated config even if it is not Odin-managed")
	cmd.Flags().BoolVar(&withMemory, "with-memory", false, "Include an explicit MemPalace MCP server entry in the generated config")
	return cmd
}

func printOpenCodeToolStatus(out io.Writer, tool plugins.Status) {
	fmt.Fprintf(out, "OpenCode CLI: %s", installedLabel(tool.Installed))
	if tool.Path != "" {
		fmt.Fprintf(out, " (%s", tool.Path)
		if tool.Details != "" {
			fmt.Fprintf(out, ", %s", tool.Details)
		}
		fmt.Fprintln(out, ")")
		return
	}
	if tool.Details != "" {
		fmt.Fprintf(out, " (%s)", tool.Details)
	}
	fmt.Fprintln(out)
}

func memoryMCPLabel(configured, enabled bool) string {
	if !configured {
		return "not configured"
	}
	if enabled {
		return "configured and enabled"
	}
	return "configured but disabled"
}
