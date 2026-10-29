package cli

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestModelRecommendRejectsUnknownUseCase(t *testing.T) {
	cmd := NewRootCmd()
	cmd.SetArgs([]string{"model", "recommend", "--use-case", "foo"})

	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), `unknown --use-case "foo"`) {
		t.Fatalf("err = %v, want unknown --use-case error", err)
	}
}

func TestModelPullFailsWhenOllamaIsUnreachable(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	if err := os.MkdirAll(filepath.Join(home, ".odin"), 0o755); err != nil {
		t.Fatal(err)
	}
	// Port 1 is reserved and refuses connections, so the probe fails fast.
	if err := os.WriteFile(filepath.Join(home, ".odin", "config.toml"), []byte("[runtime]\nbase_url = \"http://127.0.0.1:1\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	for _, args := range [][]string{{"model", "pull", "m", "--yes"}, {"model", "list"}} {
		cmd := NewRootCmd()
		cmd.SetArgs(args)
		cmd.SetOut(io.Discard)
		if err := cmd.Execute(); err == nil || !strings.Contains(err.Error(), "start") {
			t.Fatalf("odin %v: err = %v, want an error that says to start Ollama", args, err)
		}
	}
}
