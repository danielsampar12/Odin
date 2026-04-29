package cli

import "github.com/spf13/cobra"

func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "odin",
		Short:        "Manage a local-first AI workstation",
		SilenceUsage: true,
	}

	cmd.AddCommand(
		newVersionCmd(),
		newDoctorCmd(),
		newSetupCmd(),
		newInitCmd(),
		newStartCmd(),
	)

	return cmd
}

func Execute() error {
	return NewRootCmd().Execute()
}
