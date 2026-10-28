package cli

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestOpenCodeGenerateFailsOnUnmanagedConfig(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	project := t.TempDir()
	t.Chdir(project)
	if err := os.MkdirAll(filepath.Join(project, ".odin", "generated"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(project, ".odin", "generated", "opencode.jsonc"), []byte("{}\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := NewRootCmd()
	cmd.SetArgs([]string{"opencode", "generate"})
	cmd.SetOut(io.Discard)

	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "refusing to overwrite") {
		t.Fatalf("err = %v, want a refusal error", err)
	}
}
