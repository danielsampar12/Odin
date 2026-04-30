package opencode

import (
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

type LaunchOptions struct {
	BinaryPath string
	WorkingDir string
	ConfigPath string
	Env        []string
	Stdin      io.Reader
	Stdout     io.Writer
	Stderr     io.Writer
}

func RelativeConfigPath() string {
	return filepath.Join(".odin", "generated", "opencode.jsonc")
}

func LaunchCommand(binaryPath string) string {
	name := "opencode"
	if binaryPath != "" {
		name = binaryPath
	}
	return "OPENCODE_CONFIG=" + RelativeConfigPath() + " " + name + " ."
}

func Launch(ctx context.Context, options LaunchOptions) error {
	command := buildLaunchCommand(ctx, options)
	return command.Run()
}

func buildLaunchCommand(ctx context.Context, options LaunchOptions) *exec.Cmd {
	command := exec.CommandContext(ctx, options.BinaryPath, ".")
	command.Dir = options.WorkingDir
	command.Env = append(defaultEnv(options.Env), "OPENCODE_CONFIG="+options.ConfigPath)
	command.Stdin = valueOrReader(options.Stdin, os.Stdin)
	command.Stdout = valueOrWriter(options.Stdout, os.Stdout)
	command.Stderr = valueOrWriter(options.Stderr, os.Stderr)
	return command
}

func defaultEnv(env []string) []string {
	if env != nil {
		return append([]string{}, env...)
	}
	return append([]string{}, os.Environ()...)
}

func valueOrReader(value io.Reader, fallback io.Reader) io.Reader {
	if value != nil {
		return value
	}
	return fallback
}

func valueOrWriter(value io.Writer, fallback io.Writer) io.Writer {
	if value != nil {
		return value
	}
	return fallback
}
