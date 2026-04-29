package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/danielsampar12/odin/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			os.Exit(2)
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
