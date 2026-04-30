package opencode

import (
	"context"
	"io"
	"testing"
)

func TestBuildLaunchCommand(t *testing.T) {
	command := buildLaunchCommand(context.Background(), LaunchOptions{
		BinaryPath: "/usr/local/bin/opencode",
		WorkingDir: "/tmp/project",
		ConfigPath: ".odin/generated/opencode.jsonc",
		Env:        []string{"PATH=/usr/bin", "HOME=/tmp/home"},
		Stdin:      io.NopCloser(nil),
	})

	if command.Path != "/usr/local/bin/opencode" {
		t.Fatalf("command.Path = %q, want /usr/local/bin/opencode", command.Path)
	}
	if command.Dir != "/tmp/project" {
		t.Fatalf("command.Dir = %q, want /tmp/project", command.Dir)
	}
	if len(command.Args) != 2 || command.Args[1] != "." {
		t.Fatalf("command.Args = %#v, want binary plus '.'", command.Args)
	}

	found := false
	for _, env := range command.Env {
		if env == "OPENCODE_CONFIG=.odin/generated/opencode.jsonc" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("command.Env is missing OPENCODE_CONFIG: %#v", command.Env)
	}
}
