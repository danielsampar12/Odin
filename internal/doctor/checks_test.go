package doctor

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindLegacyShellFunction(t *testing.T) {
	home := t.TempDir()
	if got := findLegacyShellFunction(home); got != "" {
		t.Fatalf("empty home: got %q, want no match", got)
	}

	if err := os.WriteFile(filepath.Join(home, ".zshrc"), []byte("export PATH=$HOME/bin:$PATH\n# odin() is mentioned here\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := findLegacyShellFunction(home); got != "" {
		t.Fatalf("comment only: got %q, want no match", got)
	}

	bashrc := filepath.Join(home, ".bashrc")
	if err := os.WriteFile(bashrc, []byte("\n# odin\nodin() {\n  case \"$1\" in\n  esac\n}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := findLegacyShellFunction(home); got != bashrc {
		t.Fatalf("got %q, want %q", got, bashrc)
	}
}
