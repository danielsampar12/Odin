package cli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/danielsampar12/odin/internal/companions"
	"github.com/danielsampar12/odin/internal/doctor"
	"github.com/danielsampar12/odin/internal/models"
	"github.com/danielsampar12/odin/internal/plugins"
	mempalaceplugin "github.com/danielsampar12/odin/internal/plugins/mempalace"
	opencodeplugin "github.com/danielsampar12/odin/internal/plugins/opencode"
	"github.com/danielsampar12/odin/internal/system"
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

			numCtx := models.ContextFor(scaffold.Model, result.RAMGB, result.GPU)
			state, err := opencodeplugin.GeneratedConfigState(cwd, opencodeplugin.GenerateOptions{Context: numCtx, WithMemory: memoryStatus.OpenCodeMCPConfigured})
			if err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			fmt.Fprintln(out, "OpenCode status")
			fmt.Fprintln(out)
			printWarnings(out, scaffold.Warnings)
			printOpenCodeToolStatus(out, result.Tool("opencode"))
			fmt.Fprintf(out, "Generated config: %s (%s)\n", displayPath(scaffold.ConfigPath), state)
			fmt.Fprintf(out, "Runtime provider: %s\n", scaffold.RuntimeProvider)
			fmt.Fprintf(out, "Runtime base URL: %s\n", scaffold.RuntimeBaseURL)
			fmt.Fprintf(out, "OpenCode provider/model: %s\n", scaffold.ProviderModel(numCtx))
			fmt.Fprintf(out, "Context: %d tokens\n", numCtx)
			fmt.Fprintf(out, "Selected companion: %s (from %s)\n", companions.DisplayName(scaffold.Companion), scaffold.CompanionSource)
			fmt.Fprintf(out, "OpenCode agents: %s\n", agentSummary(scaffold.Companion))
			fmt.Fprintf(out, "Memory provider: %s\n", scaffold.MemoryProvider)
			fmt.Fprintf(out, "Memory hall: %s\n", scaffold.MemoryHall)
			fmt.Fprintf(out, "MemPalace MCP wiring: %s\n", memoryMCPLabel(memoryStatus.OpenCodeMCPConfigured, memoryStatus.OpenCodeMCPEnabled))
			printProvenance(out, cwd)
			if scaffold.Supported() {
				fmt.Fprintf(out, "Planned launch command: %s\n", opencodeplugin.LaunchCommand(result.Tool("opencode").Path))
			} else {
				fmt.Fprintf(out, "OpenCode generation currently supports Ollama only. Current runtime provider: %s\n", scaffold.RuntimeProvider)
			}
			if hint := configStateHint(state); hint != "" {
				fmt.Fprintln(out, hint)
			} else if scaffold.SupportsMemory() && !memoryStatus.OpenCodeMCPConfigured {
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

			scaffold, err := opencodeplugin.ResolveProjectScaffold(cwd)
			if err != nil {
				return err
			}
			numCtx := models.ContextFor(scaffold.Model, system.DetectRAMGB(cmd.Context()), system.DetectGPU(cmd.Context()))

			result, err := opencodeplugin.WriteGeneratedConfig(cwd, opencodeplugin.GenerateOptions{
				Force:      force,
				WithMemory: withMemory,
				Context:    numCtx,
			})
			if err != nil {
				if errors.Is(err, opencodeplugin.ErrUnmanagedGeneratedConfig) {
					return fmt.Errorf("refusing to overwrite %s because it is not Odin-managed; re-run with --force to replace it", displayPath(result.Path))
				}
				if errors.Is(err, opencodeplugin.ErrEditedGeneratedConfig) {
					return fmt.Errorf("refusing to overwrite %s because it was edited by hand; move your changes to opencode.json or re-run with --force to discard them", displayPath(result.Path))
				}
				return err
			}

			out := cmd.OutOrStdout()
			printWarnings(out, scaffold.Warnings)
			if result.Written && result.Updated {
				fmt.Fprintf(out, "Updated %s\n", displayPath(result.Path))
			} else if result.Written {
				fmt.Fprintf(out, "Created %s\n", displayPath(result.Path))
			} else {
				fmt.Fprintf(out, "Kept existing %s\n", displayPath(result.Path))
			}
			fmt.Fprintf(out, "OpenCode provider/model: %s\n", scaffold.ProviderModel(numCtx))
			fmt.Fprintf(out, "Context: %d tokens (`odin start --exec` creates Ollama model %s if it is missing)\n", numCtx, scaffold.ContextModel(numCtx))
			fmt.Fprintf(out, "Selected companion: %s\n", companions.DisplayName(scaffold.Companion))
			fmt.Fprintf(out, "Memory provider: %s\n", scaffold.MemoryProvider)
			fmt.Fprintf(out, "Memory hall: %s\n", scaffold.MemoryHall)
			if withMemory {
				fmt.Fprintln(out, "MemPalace MCP: enabled in generated config")
			} else {
				fmt.Fprintln(out, "MemPalace MCP: not included by default")
			}
			fmt.Fprintln(out, "Project instructions: .odin/rules.md, plus AGENTS.md if the repository has one")
			fmt.Fprintf(out, "Use with: %s\n", opencodeplugin.LaunchCommand(""))
			printGitignoreHint(cmd.Context(), out, cwd)

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

// printProvenance shows which OpenCode configs and instruction files apply, in the order OpenCode merges them.
func printProvenance(out io.Writer, cwd string) {
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Where OpenCode settings come from (OpenCode merges these in order; later ones win):")
	sources := opencodeplugin.Sources(cwd)
	for _, source := range sources {
		switch {
		case source.Err != nil:
			fmt.Fprintf(out, "- %s: %s\n", source.Name, source.Err)
		case len(source.Keys) == 0:
			fmt.Fprintf(out, "- %s: none of the settings Odin tracks\n", source.Name)
		default:
			fmt.Fprintf(out, "- %s: %s\n", source.Name, strings.Join(source.Keys, ", "))
		}
	}
	if len(sources) == 0 {
		fmt.Fprintln(out, "- none found")
	}
	if shellConfig := os.Getenv("OPENCODE_CONFIG"); shellConfig != "" && shellConfig != opencodeplugin.RelativeConfigPath() {
		fmt.Fprintf(out, "Note: your shell sets OPENCODE_CONFIG=%s; Odin replaces it with %s when it launches OpenCode.\n", shellConfig, opencodeplugin.RelativeConfigPath())
	}
	printModelOverrides(out, sources)

	var found []string
	for _, path := range []string{"AGENTS.md", filepath.Join(".odin", "rules.md"), filepath.Join(opencodeplugin.GlobalConfigDir(), "AGENTS.md")} {
		if !filepath.IsAbs(path) {
			path = filepath.Join(cwd, path)
		}
		if _, err := os.Stat(path); err == nil {
			found = append(found, displayPath(path))
		}
	}
	if len(found) == 0 {
		found = []string{"none"}
	}
	fmt.Fprintf(out, "Instruction files: %s, plus any \"instructions\" listed above\n", strings.Join(found, ", "))
	fmt.Fprintln(out, "Skipped at launch: CLAUDE.md files and skills in ~/.claude and ~/.agents (Odin sets OPENCODE_DISABLE_CLAUDE_CODE=1 and OPENCODE_DISABLE_EXTERNAL_SKILLS=1)")
	fmt.Fprintln(out)
}

// agentSummary lists the generated OpenCode agents, e.g. "baldur (starts by default), tyr (read-only)".
func agentSummary(defaultKey string) string {
	var parts []string
	for _, companion := range companions.Registry {
		var notes []string
		if companion.Key == defaultKey {
			notes = append(notes, "starts by default")
		}
		if companion.ReadOnly {
			notes = append(notes, "read-only")
		}
		if len(notes) > 0 {
			parts = append(parts, fmt.Sprintf("%s (%s)", companion.Key, strings.Join(notes, ", ")))
		} else {
			parts = append(parts, companion.Key)
		}
	}
	return strings.Join(parts, ", ") + "; press Tab in OpenCode to switch"
}

func printModelOverrides(out io.Writer, sources []opencodeplugin.Source) {
	for _, source := range opencodeplugin.ModelOverrides(sources) {
		fmt.Fprintf(out, "Warning: %s sets \"model\": %q, which overrides the model in Odin's generated config.\n", source.Name, source.Model)
	}
}

// configStateHint tells the user what to do about a generated config that is not up to date.
func configStateHint(state opencodeplugin.ConfigState) string {
	switch state {
	case opencodeplugin.ConfigMissing:
		return "Generated config is missing. Run `odin opencode generate`."
	case opencodeplugin.ConfigOutdated:
		return "Generated config is out of date for the current settings. Re-run `odin opencode generate`."
	case opencodeplugin.ConfigEdited:
		return "Generated config was edited by hand; `odin opencode generate` will not overwrite it without --force. Put project changes in opencode.json instead."
	case opencodeplugin.ConfigUnmanaged:
		return "Generated config was not written by Odin. Run `odin opencode generate --force` to replace it."
	}
	return ""
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
