package cli

import (
	"encoding/json"
	"flag"
	"fmt"

	"github.com/danielsampar12/odin/internal/companions"
	opencodeplugin "github.com/danielsampar12/odin/internal/plugins/opencode"
	"github.com/spf13/cobra"
)

func newCompanionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "companion",
		Short: "List and inspect companions (OpenCode agents Odin generates)",
	}
	cmd.AddCommand(newCompanionListCmd(), newCompanionShowCmd())
	return cmd
}

func newCompanionListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List companions and show which one OpenCode starts with",
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := workingDir()
			if err != nil {
				return err
			}
			scaffold, err := opencodeplugin.ResolveProjectScaffold(cwd)
			if err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			printWarnings(out, scaffold.Warnings)
			for _, companion := range companions.Registry {
				marker := " "
				if companion.Key == scaffold.Companion {
					marker = "*"
				}
				fmt.Fprintf(out, "%s %-7s %s\n", marker, companion.Key, companion.Description)
			}
			fmt.Fprintln(out)
			fmt.Fprintln(out, "* starts by default. Choose with [companion] default in .odin/config.toml (this project) or ~/.odin/config.toml, then run `odin opencode generate`.")
			fmt.Fprintln(out, "In OpenCode, press Tab to switch companions, or run `opencode run --agent <name>`.")
			return nil
		},
	}
}

func newCompanionShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show <name>",
		Short: "Show a companion's permissions and prompt",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 1 {
				return nil
			}
			cmd.PrintErrln("Error: expected exactly 1 companion name")
			return flag.ErrHelp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			companion, ok := companions.Lookup(args[0])
			if !ok {
				return fmt.Errorf("unknown companion %q; run `odin companion list`", args[0])
			}

			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "%s (%s)\n", companion.Name, companion.Key)
			fmt.Fprintln(out, companion.Description)
			agent := opencodeplugin.AgentConfig(companion)
			prompt := agent["prompt"]
			delete(agent, "prompt")
			body, err := json.MarshalIndent(agent, "", "  ")
			if err != nil {
				return err
			}
			fmt.Fprintf(out, "OpenCode agent %q, as `odin opencode generate` writes it (without the prompt):\n%s\n", companion.Key, body)
			fmt.Fprintln(out)
			fmt.Fprintln(out, "Prompt (replaces OpenCode's built-in prompt for this agent):")
			fmt.Fprintln(out)
			fmt.Fprint(out, prompt)
			return nil
		},
	}
}
