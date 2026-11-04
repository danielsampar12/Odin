package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadReadsTOMLTheOldParserMissed(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	body := `profile = 'developer' # a "quoted" comment
runtime.base_url = "http://gpu-box:11434"
model = { default = "qwen3.5:9b-q4_K_M" }

[companion]
default="tyr"
`
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	file, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if file.Profile != "developer" || file.Runtime.BaseURL != "http://gpu-box:11434" || file.Model.Default != "qwen3.5:9b-q4_K_M" || file.Companion.Default != "tyr" {
		t.Fatalf("Load = %+v", file)
	}
}

func TestLoadMissingAndInvalid(t *testing.T) {
	dir := t.TempDir()
	if file, err := Load(filepath.Join(dir, "missing.toml")); err != nil || file != (File{}) {
		t.Fatalf("Load(missing) = %+v, %v; want an empty File and no error", file, err)
	}

	path := filepath.Join(dir, "bad.toml")
	if err := os.WriteFile(path, []byte("[model\ndefault = \"x\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(path); err == nil || !strings.Contains(err.Error(), path) {
		t.Fatalf("Load(invalid) error = %v, want one naming the file", err)
	}
}

func TestDefaultConfigsRoundTrip(t *testing.T) {
	dir := t.TempDir()
	project := filepath.Join(dir, "project.toml")
	if err := os.WriteFile(project, []byte(DefaultProjectConfig(ProjectSettings{Name: "demo", MemoryProvider: "none", MemoryHall: "demo", CompanionDefault: "baldur"})), 0o644); err != nil {
		t.Fatal(err)
	}
	file, err := Load(project)
	if err != nil {
		t.Fatal(err)
	}
	if file.Name != "demo" || file.Companion.Default != "baldur" || file.Model.Default != "" {
		t.Fatalf("project config = %+v; want shared intent only (no model)", file)
	}

	global := filepath.Join(dir, "global.toml")
	if err := os.WriteFile(global, []byte(DefaultGlobalConfig(GlobalSettings{RuntimeBaseURL: "http://localhost:11434", ModelDefault: "m", ShellEnabled: true})), 0o644); err != nil {
		t.Fatal(err)
	}
	if file, err = Load(global); err != nil || file.Model.Default != "m" || file.Runtime.BaseURL != "http://localhost:11434" || !file.Shell.Enabled {
		t.Fatalf("global config = %+v, %v", file, err)
	}
}
