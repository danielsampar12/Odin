package setupplan

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/danielsampar12/odin/internal/doctor"
	"github.com/danielsampar12/odin/internal/system"
)

type fakeRunner struct {
	commands []string
	failures map[string]error
}

func (f *fakeRunner) Run(_ context.Context, command string, _ RunOptions) error {
	f.commands = append(f.commands, command)
	if f.failures != nil {
		if err, ok := f.failures[command]; ok {
			return err
		}
	}
	return nil
}

func TestExecuteCreatesGlobalConfigAndLeavesManualSteps(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	plan := Plan{
		Recommendation: Recommendation{
			Profile:        "developer",
			Model:          "qwen2.5-coder:7b",
			FallbackModel:  "qwen2.5-coder:3b",
			CompanionKey:   "baldur",
			ShellProvider:  "p10k",
			RuntimeBaseURL: "http://localhost:11434",
		},
		Steps: []Step{
			{
				ID:        StepGlobalConfig,
				Name:      "Create ~/.odin/config.toml",
				Status:    StatusNeeded,
				Automatic: true,
			},
			{
				ID:      StepInstallOllama,
				Name:    "Install Ollama",
				Status:  StatusManual,
				Command: "curl -fsSL https://ollama.com/install.sh | sh",
			},
		},
	}

	result, err := Execute(context.Background(), t.TempDir(), plan, ExecuteOptions{
		Runner: &fakeRunner{},
	})
	if err != nil {
		t.Fatalf("Execute error = %v", err)
	}

	if _, err := os.Stat(filepath.Join(home, ".odin", "config.toml")); err != nil {
		t.Fatalf("expected global config to be created: %v", err)
	}

	completed, skipped, failed, manual := result.Counts()
	if completed != 1 || skipped != 0 || failed != 0 || manual != 1 {
		t.Fatalf("counts = (%d,%d,%d,%d), want (1,0,0,1)", completed, skipped, failed, manual)
	}
}

func TestExecuteStopsOnCriticalOpenCodeFailure(t *testing.T) {
	runner := &fakeRunner{
		failures: map[string]error{
			"npm install -g opencode-ai": errors.New("boom"),
		},
	}

	plan := Plan{
		Steps: []Step{
			{
				ID:        StepInstallOpenCode,
				Name:      "Install OpenCode",
				Status:    StatusNeeded,
				Command:   "npm install -g opencode-ai",
				Automatic: true,
			},
			{
				ID:        StepInstallMemPalace,
				Name:      "Install MemPalace",
				Status:    StatusNeeded,
				Command:   "pip install mempalace",
				Automatic: true,
			},
		},
	}

	result, err := Execute(context.Background(), t.TempDir(), plan, ExecuteOptions{
		Runner: runner,
	})
	if err != nil {
		t.Fatalf("Execute error = %v", err)
	}

	if len(result.Steps) != 1 {
		t.Fatalf("executed %d steps, want 1 after critical stop", len(result.Steps))
	}
	if !result.HasCriticalFailure() {
		t.Fatal("expected critical failure")
	}
	if len(runner.commands) != 1 {
		t.Fatalf("runner commands = %d, want 1", len(runner.commands))
	}
}

func TestExecuteMemPalaceFailureContinues(t *testing.T) {
	runner := &fakeRunner{
		failures: map[string]error{
			"pip install mempalace": errors.New("boom"),
		},
	}

	plan := Plan{
		Result: doctor.Result{
			Ollama: doctor.Result{}.Ollama,
			OS:     system.OSInfo{Name: "linux"},
		},
		Steps: []Step{
			{
				ID:        StepInstallMemPalace,
				Name:      "Install MemPalace",
				Status:    StatusNeeded,
				Command:   "pip install mempalace",
				Automatic: true,
			},
			{
				ID:     StepShellIntegration,
				Name:   "Shell integration",
				Status: StatusSkipped,
			},
		},
	}

	result, err := Execute(context.Background(), t.TempDir(), plan, ExecuteOptions{
		Runner: runner,
	})
	if err != nil {
		t.Fatalf("Execute error = %v", err)
	}

	if len(result.Steps) != 2 {
		t.Fatalf("executed %d steps, want 2", len(result.Steps))
	}
	if result.HasCriticalFailure() {
		t.Fatal("did not expect critical failure")
	}
}
