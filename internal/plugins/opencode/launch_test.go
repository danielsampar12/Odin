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

	for _, want := range []string{"OPENCODE_CONFIG=.odin/generated/opencode.jsonc", "OPENCODE_DISABLE_CLAUDE_CODE=1", "OPENCODE_DISABLE_EXTERNAL_SKILLS=1"} {
		found := false
		for _, env := range command.Env {
			found = found || env == want
		}
		if !found {
			t.Fatalf("command.Env is missing %s: %#v", want, command.Env)
		}
	}
}

func TestLaunchCommandKeepsGlobalInstructionsOut(t *testing.T) {
	got := LaunchCommand("")
	want := "OPENCODE_CONFIG=" + RelativeConfigPath() + " OPENCODE_DISABLE_CLAUDE_CODE=1 OPENCODE_DISABLE_EXTERNAL_SKILLS=1 opencode ."
	if got != want {
		t.Fatalf("LaunchCommand = %q, want %q", got, want)
	}
}
