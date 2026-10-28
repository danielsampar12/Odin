package mempalace

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/danielsampar12/odin/internal/config"
)

func TestResolveProjectStatus(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	projectDir := t.TempDir()
	if err := os.MkdirAll(config.ProjectDir(projectDir), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(config.GlobalDir(), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(GlobalDir(), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(config.GlobalConfigPath(), []byte(`[memory]
provider = "mempalace"
`), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(config.ProjectConfigPath(projectDir), []byte(`[memory]
provider = "mempalace"
hall = "demo-hall"
`), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(GlobalConfigPath(), []byte(`{"palace_path":"/tmp/custom-palace"}`), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := os.MkdirAll(config.ProjectGeneratedDir(projectDir), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(config.ProjectGeneratedOpenCodeConfigPath(projectDir), []byte(`{
  "mcp": {
    "mempalace": {
      "type": "local",
      "command": ["python", "-m", "mempalace.mcp_server"],
      "enabled": true
    }
  }
}`), 0o644); err != nil {
		t.Fatal(err)
	}

	status, err := ResolveProjectStatus(projectDir)
	if err != nil {
		t.Fatalf("ResolveProjectStatus error = %v", err)
	}

	if status.ResolvedProvider != ExpectedProvider {
		t.Fatalf("ResolvedProvider = %q, want %q", status.ResolvedProvider, ExpectedProvider)
	}
	if status.Hall != "demo-hall" {
		t.Fatalf("Hall = %q, want demo-hall", status.Hall)
	}
	if status.HallDerived {
		t.Fatalf("HallDerived = true, want false")
	}
	if status.PalacePath != "/tmp/custom-palace" {
		t.Fatalf("PalacePath = %q, want /tmp/custom-palace", status.PalacePath)
	}
	if !status.OpenCodeMCPConfigured || !status.OpenCodeMCPEnabled {
		t.Fatalf("expected OpenCode MCP wiring to be detected")
	}
}

func TestDetectOpenCodeMCPMissingFile(t *testing.T) {
	_, err := DetectOpenCodeMCP(filepath.Join(t.TempDir(), "missing-opencode.jsonc"))
	if err == nil {
		t.Fatal("expected missing file error")
	}
}
