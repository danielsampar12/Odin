package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Version is replaced with the release tag at build time:
// -ldflags "-X github.com/danielsampar12/odin/internal/cli.Version=v0.2.0"
var Version = "v0.3.0-dev"

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the Odin version",
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := fmt.Fprintf(cmd.OutOrStdout(), "odin %s\n", Version)
			return err
		},
	}
}
