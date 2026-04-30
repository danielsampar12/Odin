package setupplan

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/danielsampar12/odin/internal/doctor"
	"github.com/danielsampar12/odin/internal/plugins"
	ollamaplugin "github.com/danielsampar12/odin/internal/plugins/ollama"
	"github.com/danielsampar12/odin/internal/system"
)

func TestPlanOpenCodeInstallPrefersNPM(t *testing.T) {
	method := planOpenCodeInstall(doctor.Result{
		OS: system.OSInfo{Name: "linux"},
	}, Environment{
		NPM:  system.CommandStatus{Name: "npm", Installed: true, Path: "/usr/bin/npm"},
		Bun:  system.CommandStatus{Name: "bun", Installed: true, Path: "/usr/bin/bun"},
		Brew: system.CommandStatus{Name: "brew", Installed: true, Path: "/usr/bin/brew"},
	})

	if method.Command != "npm install -g opencode-ai" {
		t.Fatalf("Command = %q, want npm install -g opencode-ai", method.Command)
	}
	if !method.Automatic {
		t.Fatalf("Automatic = false, want true")
	}
}

func TestPlanMemPalaceInstallNeedsDocumentedPip(t *testing.T) {
	method := planMemPalaceInstall(doctor.Result{
		OS: system.OSInfo{Name: "linux"},
	}, Environment{
		Python3: system.CommandStatus{Name: "python3", Installed: true, Path: "/usr/bin/python3"},
		Pip3:    system.CommandStatus{Name: "pip3", Installed: true, Path: "/usr/bin/pip3"},
	})

	if method.Automatic {
		t.Fatalf("Automatic = true, want false when only pip3 is available")
	}
	if method.Command != "Install Python 3.9+ with pip, then run: pip install mempalace" {
		t.Fatalf("unexpected command: %q", method.Command)
	}
}

func TestBuildStepsModelPullNeededWhenOllamaAPIReady(t *testing.T) {
	result := doctor.Result{
		CurrentDir: "/tmp/project",
		OS:         system.OSInfo{Name: "linux"},
		Ollama: ollamaplugin.APIStatus{
			BaseURL:      ollamaplugin.DefaultBaseURL,
			APIAvailable: true,
			Models: []ollamaplugin.Model{
				{Name: "different-model"},
			},
		},
		Tools: map[string]plugins.Status{
			"ollama": {Name: "Ollama", Installed: true, Path: "/usr/bin/ollama"},
		},
	}

	recommendation := Recommendation{
		Model:          "qwen2.5-coder:7b",
		CompanionName:  "Baldur",
		Profile:        "developer",
		RuntimeBaseURL: ollamaplugin.DefaultBaseURL,
	}

	steps := buildSteps(result, recommendation, InstallMethod{}, InstallMethod{}, InstallMethod{}, InstallMethod{})

	found := false
	for _, step := range steps {
		if step.Name != "Pull recommended model" {
			continue
		}
		found = true
		if step.Status != StatusNeeded {
			t.Fatalf("status = %q, want %q", step.Status, StatusNeeded)
		}
		if step.Command != "odin model pull qwen2.5-coder:7b" {
			t.Fatalf("command = %q", step.Command)
		}
	}

	if !found {
		t.Fatal("did not find model pull step")
	}
}

func TestBuildDoesNotCreateGlobalConfig(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	cwd := t.TempDir()
	if _, err := Build(context.Background(), cwd); err != nil {
		t.Fatalf("Build error = %v", err)
	}

	if _, err := os.Stat(filepath.Join(home, ".odin")); !os.IsNotExist(err) {
		t.Fatalf("expected Build to stay read-only, stat err = %v", err)
	}
}
