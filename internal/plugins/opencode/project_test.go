package opencode

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/danielsampar12/odin/internal/config"
)

func TestWriteGeneratedConfig(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	projectDir := t.TempDir()
	if err := os.MkdirAll(config.ProjectDir(projectDir), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(config.GlobalDir(), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(config.GlobalConfigPath(), []byte(`profile = "developer"

[runtime]
provider = "ollama"
base_url = "http://localhost:11434"

[model]
default = "global-model"

[companion]
default = "freya"
`), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(config.ProjectConfigPath(projectDir), []byte(`name = "demo"

[runtime]
provider = "ollama"

[model]
default = "project-model"

[companion]
default = "baldur"
`), 0o644); err != nil {
		t.Fatal(err)
	}

	result, err := WriteGeneratedConfig(projectDir, false)
	if err != nil {
		t.Fatalf("WriteGeneratedConfig error = %v", err)
	}
	if !result.Written {
		t.Fatalf("expected config to be written")
	}

	body, err := os.ReadFile(config.ProjectGeneratedOpenCodeConfigPath(projectDir))
	if err != nil {
		t.Fatal(err)
	}
	content := string(body)
	if !strings.Contains(content, ManagedMarker) {
		t.Fatalf("generated config is missing managed marker")
	}
	if !strings.Contains(content, `"model": "ollama/project-model"`) {
		t.Fatalf("generated config is missing provider/model reference: %s", content)
	}
	if !strings.Contains(content, `"baseURL": "http://localhost:11434/v1"`) {
		t.Fatalf("generated config is missing Ollama v1 base URL: %s", content)
	}
	if !strings.Contains(content, `"instructions": [`) || !strings.Contains(content, "../rules.md") {
		t.Fatalf("generated config is missing rules instruction reference: %s", content)
	}
}

func TestWriteGeneratedConfigRefusesUnmanagedFile(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	projectDir := t.TempDir()
	if err := os.MkdirAll(config.ProjectGeneratedDir(projectDir), 0o755); err != nil {
		t.Fatal(err)
	}

	path := config.ProjectGeneratedOpenCodeConfigPath(projectDir)
	if err := os.WriteFile(path, []byte(`{"custom":true}`), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := WriteGeneratedConfig(projectDir, false)
	if !errors.Is(err, ErrUnmanagedGeneratedConfig) {
		t.Fatalf("error = %v, want ErrUnmanagedGeneratedConfig", err)
	}

	result, err := WriteGeneratedConfig(projectDir, true)
	if err != nil {
		t.Fatalf("force WriteGeneratedConfig error = %v", err)
	}
	if !result.Written {
		t.Fatalf("expected forced regeneration to write file")
	}
}

func TestEnsureV1Endpoint(t *testing.T) {
	testCases := []struct {
		input string
		want  string
	}{
		{input: "http://localhost:11434", want: "http://localhost:11434/v1"},
		{input: "http://localhost:11434/", want: "http://localhost:11434/v1"},
		{input: "http://localhost:11434/v1", want: "http://localhost:11434/v1"},
	}

	for _, tc := range testCases {
		if got := ensureV1Endpoint(tc.input); got != tc.want {
			t.Fatalf("ensureV1Endpoint(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}
