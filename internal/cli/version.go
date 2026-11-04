package cli

import (
	"fmt"
	"runtime/debug"

	"github.com/spf13/cobra"
)

// Version is replaced with the release tag at build time:
// -ldflags "-X github.com/danielsampar12/odin/internal/cli.Version=v0.2.0"
var Version string

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the Odin version",
		RunE: func(cmd *cobra.Command, args []string) error {
			var module string
			if info, ok := debug.ReadBuildInfo(); ok {
				module = info.Main.Version
			}
			_, err := fmt.Fprintf(cmd.OutOrStdout(), "odin %s\n", displayVersion(Version, module))
			return err
		},
	}
}

// displayVersion falls back to the module version Go records for `go install` builds, which have no release tag.
func displayVersion(release, module string) string {
	if release != "" {
		return release
	}
	if module != "" && module != "(devel)" {
		return module
	}
	return "dev"
}
