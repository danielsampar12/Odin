package cli

import (
	"fmt"

	"github.com/danielsampar12/odin/internal/doctor"
	mempalaceplugin "github.com/danielsampar12/odin/internal/plugins/mempalace"
	"github.com/spf13/cobra"
)

func newMemoryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "memory",
		Short: "Inspect MemPalace memory integration",
	}

	cmd.AddCommand(
		newMemoryStatusCmd(),
		newMemoryDoctorCmd(),
	)

	return cmd
}

func newMemoryStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show memory integration status for the current project",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMemoryReport(cmd, false)
		},
	}
}

func newMemoryDoctorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Diagnose MemPalace and Odin memory integration",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMemoryReport(cmd, true)
		},
	}
}

func runMemoryReport(cmd *cobra.Command, verbose bool) error {
	cwd, err := workingDir()
	if err != nil {
		return err
	}

	result, err := doctor.Collect(cmd.Context(), cwd)
	if err != nil {
		return err
	}

	status, err := mempalaceplugin.ResolveProjectStatus(cwd)
	if err != nil {
		return err
	}

	out := cmd.OutOrStdout()
	title := "Odin memory status"
	if verbose {
		title = "Odin memory doctor"
	}

	fmt.Fprintln(out, title)
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Config")
	fmt.Fprintf(out, "- Expected provider: %s\n", mempalaceplugin.ExpectedProvider)
	fmt.Fprintf(out, "- Global provider: %s\n", configValueOrUnset(status.GlobalProvider))
	fmt.Fprintf(out, "- Project provider: %s\n", configValueOrUnset(status.ProjectProvider))
	fmt.Fprintf(out, "- Resolved provider: %s\n", configValueOrUnset(status.ResolvedProvider))
	fmt.Fprintf(out, "- Project config: %s\n", fileStatusLabel(status.ProjectConfigPath, status.ProjectConfigExists))
	fmt.Fprintf(out, "- Project hall: %s%s\n", status.Hall, derivedSuffix(status.HallDerived))
	fmt.Fprintln(out)
	fmt.Fprintln(out, "MemPalace")
	printToolStatusLine(out, "MemPalace CLI", result.Tool("mempalace"))
	fmt.Fprintf(out, "- Palace config: %s\n", fileStatusLabel(status.PalaceConfigPath, status.PalaceConfigExists))
	fmt.Fprintf(out, "- Palace identity: %s\n", fileStatusLabel(status.PalaceIdentityPath, status.PalaceIdentityExists))
	fmt.Fprintf(out, "- Palace path: %s\n", displayPath(status.PalacePath))
	if verbose {
		fmt.Fprintln(out, "- Storage: MemPalace stores drawers locally on disk and uses a local knowledge graph.")
		fmt.Fprintf(out, "- Documented MCP command: %s\n", joinCommand(mempalaceplugin.MCPCommand()))
	}
	fmt.Fprintln(out)
	fmt.Fprintln(out, "OpenCode")
	fmt.Fprintf(out, "- Generated config: %s\n", fileStatusLabel(status.OpenCodeConfigPath, status.OpenCodeConfigExists))
	fmt.Fprintf(out, "- MemPalace MCP wiring: %s\n", memoryMCPLabel(status.OpenCodeMCPConfigured, status.OpenCodeMCPEnabled))
	if status.OpenCodeConfigExists && !status.OpenCodeMCPConfigured {
		fmt.Fprintln(out, "- Opt in with: odin opencode generate --with-memory")
	}
	if verbose && status.OpenCodeMCPConfigured && status.OpenCodeMCPCommandKnown {
		fmt.Fprintf(out, "- MCP command recognized: %s\n", joinCommand(mempalaceplugin.MCPCommand()))
	}

	return nil
}
