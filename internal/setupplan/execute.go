package setupplan

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/danielsampar12/odin/internal/config"
	"github.com/danielsampar12/odin/internal/plugins/mempalace"
	ollamaplugin "github.com/danielsampar12/odin/internal/plugins/ollama"
	opencodeplugin "github.com/danielsampar12/odin/internal/plugins/opencode"
)

type Outcome string

const (
	OutcomeCompleted Outcome = "completed"
	OutcomeSkipped   Outcome = "skipped"
	OutcomeFailed    Outcome = "failed"
	OutcomeManual    Outcome = "manual"
)

type StepResult struct {
	Step     Step
	Outcome  Outcome
	Message  string
	Executed bool
	Critical bool
	Err      error
}

type ExecutionResult struct {
	Steps []StepResult
}

type ExecuteOptions struct {
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
	Runner Runner
}

type Runner interface {
	Run(ctx context.Context, command string, options RunOptions) error
}

type RunOptions struct {
	Dir    string
	Env    []string
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

type ShellRunner struct{}

func (ShellRunner) Run(ctx context.Context, command string, options RunOptions) error {
	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	cmd.Dir = options.Dir
	cmd.Env = append(os.Environ(), options.Env...)
	cmd.Stdin = options.Stdin
	cmd.Stdout = options.Stdout
	cmd.Stderr = options.Stderr
	return cmd.Run()
}

func Execute(ctx context.Context, cwd string, plan Plan, options ExecuteOptions) (ExecutionResult, error) {
	runner := options.Runner
	if runner == nil {
		runner = ShellRunner{}
	}

	result := ExecutionResult{}
	for _, step := range plan.Steps {
		stepResult, err := executeStep(ctx, cwd, plan, step, runner, options)
		result.Steps = append(result.Steps, stepResult)
		if err != nil {
			return result, err
		}
		if stepResult.Outcome == OutcomeFailed && stepResult.Critical {
			return result, nil
		}
	}

	return result, nil
}

func (r ExecutionResult) Counts() (completed, skipped, failed, manual int) {
	for _, step := range r.Steps {
		switch step.Outcome {
		case OutcomeCompleted:
			completed++
		case OutcomeSkipped:
			skipped++
		case OutcomeFailed:
			failed++
		case OutcomeManual:
			manual++
		}
	}

	return completed, skipped, failed, manual
}

func (r ExecutionResult) HasCriticalFailure() bool {
	for _, step := range r.Steps {
		if step.Outcome == OutcomeFailed && step.Critical {
			return true
		}
	}

	return false
}

func executeStep(ctx context.Context, cwd string, plan Plan, step Step, runner Runner, options ExecuteOptions) (StepResult, error) {
	switch step.Status {
	case StatusDone:
		return StepResult{Step: step, Outcome: OutcomeCompleted, Message: step.Reason}, nil
	case StatusSkipped:
		return StepResult{Step: step, Outcome: OutcomeSkipped, Message: step.Reason}, nil
	case StatusManual, StatusUnsupported:
		return StepResult{Step: step, Outcome: OutcomeManual, Message: manualMessage(step)}, nil
	}

	switch step.ID {
	case StepGlobalConfig:
		return executeGlobalConfigStep(plan, step)
	case StepInstallOllama:
		return executeInstallCommandStep(ctx, step, runner, options, verifyOllamaInstall, true)
	case StepVerifyOllama:
		return executeOllamaVerifyStep(ctx, plan, step)
	case StepInstallOpenCode:
		return executeInstallCommandStep(ctx, step, runner, options, verifyOpenCodeInstall, true)
	case StepInstallMemPalace:
		return executeInstallCommandStep(ctx, step, runner, options, verifyMemPalaceInstall, false)
	case StepPullModel:
		return executeModelPullStep(ctx, plan, step, options.Stdout)
	case StepRegisterCompanions, StepShellIntegration, StepProjectInit:
		return StepResult{Step: step, Outcome: OutcomeManual, Message: manualMessage(step)}, nil
	default:
		return StepResult{Step: step, Outcome: OutcomeManual, Message: manualMessage(step)}, nil
	}
}

func executeGlobalConfigStep(plan Plan, step Step) (StepResult, error) {
	globalDir := config.GlobalDir()
	if err := os.MkdirAll(globalDir, 0o755); err != nil {
		return StepResult{
			Step:     step,
			Outcome:  OutcomeFailed,
			Message:  fmt.Sprintf("Failed to create %s.", displayPathForSetup(globalDir)),
			Critical: true,
			Err:      err,
		}, nil
	}

	globalConfigPath := config.GlobalConfigPath()
	created, err := writeFileIfMissing(globalConfigPath, config.DefaultGlobalConfig(config.GlobalSettings{
		Profile:          plan.Recommendation.Profile,
		RuntimeProvider:  "ollama",
		RuntimeBaseURL:   plan.Recommendation.RuntimeBaseURL,
		AgentProvider:    "opencode",
		MemoryProvider:   "mempalace",
		ShellProvider:    plan.Recommendation.ShellProvider,
		ShellEnabled:     plan.Recommendation.ShellEnabled,
		ModelDefault:     plan.Recommendation.Model,
		ModelFallback:    plan.Recommendation.FallbackModel,
		CompanionDefault: plan.Recommendation.CompanionKey,
	}), 0o644)
	if err != nil {
		return StepResult{
			Step:     step,
			Outcome:  OutcomeFailed,
			Message:  fmt.Sprintf("Failed to write %s.", displayPathForSetup(globalConfigPath)),
			Critical: true,
			Err:      err,
		}, nil
	}

	message := fmt.Sprintf("Kept existing %s.", displayPathForSetup(globalConfigPath))
	if created {
		message = fmt.Sprintf("Created %s.", displayPathForSetup(globalConfigPath))
	}

	return StepResult{
		Step:     step,
		Outcome:  OutcomeCompleted,
		Message:  message,
		Executed: true,
	}, nil
}

func executeInstallCommandStep(ctx context.Context, step Step, runner Runner, options ExecuteOptions, verify func() error, critical bool) (StepResult, error) {
	if !step.Automatic || strings.TrimSpace(step.Command) == "" {
		return StepResult{Step: step, Outcome: OutcomeManual, Message: manualMessage(step)}, nil
	}

	if options.Stdout != nil {
		fmt.Fprintf(options.Stdout, "Running %s\n", step.Name)
		fmt.Fprintf(options.Stdout, "Command: %s\n", step.Command)
	}

	err := runner.Run(ctx, step.Command, RunOptions{
		Stdin:  options.Stdin,
		Stdout: options.Stdout,
		Stderr: options.Stderr,
	})
	if err != nil {
		return StepResult{
			Step:     step,
			Outcome:  OutcomeFailed,
			Message:  fmt.Sprintf("%s failed.", step.Name),
			Executed: true,
			Critical: critical,
			Err:      err,
		}, nil
	}

	if verify != nil {
		if err := verify(); err != nil {
			return StepResult{
				Step:     step,
				Outcome:  OutcomeFailed,
				Message:  fmt.Sprintf("%s completed, but verification failed.", step.Name),
				Executed: true,
				Critical: critical,
				Err:      err,
			}, nil
		}
	}

	return StepResult{
		Step:     step,
		Outcome:  OutcomeCompleted,
		Message:  fmt.Sprintf("%s completed successfully.", step.Name),
		Executed: true,
	}, nil
}

func executeOllamaVerifyStep(ctx context.Context, plan Plan, step Step) (StepResult, error) {
	if plan.Result.Ollama.APIAvailable {
		return StepResult{Step: step, Outcome: OutcomeCompleted, Message: step.Reason}, nil
	}

	if !plan.Result.Tool("ollama").Installed {
		return StepResult{Step: step, Outcome: OutcomeManual, Message: manualMessage(step)}, nil
	}

	apiStatus := ProbeOllamaAPIWithRetry(ctx, plan.Result.Ollama.BaseURL, 3, 2*time.Second)
	if apiStatus.APIAvailable {
		return StepResult{
			Step:     step,
			Outcome:  OutcomeCompleted,
			Message:  fmt.Sprintf("Ollama API is responding at %s.", apiStatus.BaseURL),
			Executed: true,
		}, nil
	}

	message := fmt.Sprintf("Ollama is installed, but the API at %s is still unavailable. Start it manually with `ollama serve` or your system service manager.", apiStatus.BaseURL)
	return StepResult{Step: step, Outcome: OutcomeManual, Message: message}, nil
}

func executeModelPullStep(ctx context.Context, plan Plan, step Step, stdout io.Writer) (StepResult, error) {
	if step.Status != StatusNeeded {
		return StepResult{Step: step, Outcome: OutcomeSkipped, Message: step.Reason}, nil
	}

	if !plan.Result.Ollama.APIAvailable {
		return StepResult{Step: step, Outcome: OutcomeSkipped, Message: step.Reason}, nil
	}

	if stdout != nil {
		fmt.Fprintf(stdout, "Pulling model %s with Ollama\n", plan.Recommendation.Model)
	}

	response, err := ollamaplugin.PullModel(ctx, plan.Result.Ollama.BaseURL, ollamaplugin.PullRequest{
		Model: plan.Recommendation.Model,
	})
	if err != nil {
		return StepResult{
			Step:     step,
			Outcome:  OutcomeFailed,
			Message:  fmt.Sprintf("Model pull failed for %s.", plan.Recommendation.Model),
			Executed: true,
			Critical: false,
			Err:      err,
		}, nil
	}

	message := fmt.Sprintf("Pulled model %s.", plan.Recommendation.Model)
	if response.Status != "" {
		message = fmt.Sprintf("Pulled model %s (%s).", plan.Recommendation.Model, response.Status)
	}

	return StepResult{
		Step:     step,
		Outcome:  OutcomeCompleted,
		Message:  message,
		Executed: true,
	}, nil
}

func verifyOllamaInstall() error {
	status := ollamaplugin.Detect()
	if !status.Installed {
		return errors.New("`ollama` is still not available on PATH")
	}
	return nil
}

func verifyOpenCodeInstall() error {
	status := opencodeplugin.Detect()
	if !status.Installed {
		return errors.New("`opencode` is still not available on PATH")
	}
	if !opencodeplugin.Working(status) {
		if status.Details != "" {
			return errors.New(status.Details)
		}
		return errors.New("OpenCode binary did not pass version verification")
	}
	return nil
}

func verifyMemPalaceInstall() error {
	status := mempalace.Detect()
	if !status.Installed {
		return errors.New("`mempalace` is still not available on PATH")
	}
	if !mempalace.Working(status) {
		if status.Details != "" {
			return errors.New(status.Details)
		}
		return errors.New("MemPalace binary did not pass MCP helper verification")
	}
	return nil
}

func manualMessage(step Step) string {
	parts := []string{step.Reason}
	if step.Command != "" {
		parts = append(parts, "Suggested command: "+step.Command)
	}
	if step.Notes != "" {
		parts = append(parts, step.Notes)
	}
	return strings.Join(parts, " ")
}

func writeFileIfMissing(path, content string, perm os.FileMode) (bool, error) {
	if _, err := os.Stat(path); err == nil {
		return false, nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return false, err
	}

	return true, os.WriteFile(path, []byte(content), perm)
}

func displayPathForSetup(path string) string {
	home, err := os.UserHomeDir()
	if err == nil && home != "" && strings.HasPrefix(path, home) {
		return "~" + strings.TrimPrefix(path, home)
	}

	if cwd, err := os.Getwd(); err == nil {
		if rel, relErr := filepath.Rel(cwd, path); relErr == nil && rel != ".." && !strings.HasPrefix(rel, "../") {
			return rel
		}
	}

	return path
}

func ProbeOllamaAPIWithRetry(ctx context.Context, baseURL string, attempts int, delay time.Duration) ollamaplugin.APIStatus {
	if attempts < 1 {
		attempts = 1
	}
	if delay < 0 {
		delay = 0
	}

	var status ollamaplugin.APIStatus
	for attempt := 0; attempt < attempts; attempt++ {
		status = ollamaplugin.Probe(ctx, baseURL)
		if status.APIAvailable {
			return status
		}
		if attempt == attempts-1 || delay == 0 {
			break
		}

		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return status
		case <-timer.C:
		}
	}

	return status
}
