package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"

	"github.com/danielsampar12/odin/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			os.Exit(2)
		}
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			os.Exit(exitErr.ExitCode())
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
