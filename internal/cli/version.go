package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

const Version = "v0.2.0-dev"

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
