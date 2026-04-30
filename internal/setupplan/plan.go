package setupplan

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/danielsampar12/odin/internal/companions"
	"github.com/danielsampar12/odin/internal/doctor"
	"github.com/danielsampar12/odin/internal/models"
	ollamaplugin "github.com/danielsampar12/odin/internal/plugins/ollama"
	"github.com/danielsampar12/odin/internal/system"
)

type StepStatus string
type StepID string

const (
	StatusDone        StepStatus = "done"
	StatusNeeded      StepStatus = "needed"
	StatusSkipped     StepStatus = "skipped"
	StatusUnsupported StepStatus = "unsupported"
	StatusManual      StepStatus = "manual"
)

const (
	StepCheckSystem        StepID = "check_system"
	StepGlobalConfig       StepID = "global_config"
	StepInstallOllama      StepID = "install_ollama"
	StepVerifyOllama       StepID = "verify_ollama"
	StepInstallOpenCode    StepID = "install_opencode"
	StepInstallMemPalace   StepID = "install_mempalace"
	StepPullModel          StepID = "pull_model"
	StepRegisterCompanions StepID = "register_companions"
	StepShellIntegration   StepID = "shell_integration"
	StepProjectInit        StepID = "project_init"
)

type Recommendation struct {
	Profile        string
	Model          string
	FallbackModel  string
	Agent          string
	Memory         string
	CompanionKey   string
	CompanionName  string
	ShellProvider  string
	ShellEnabled   bool
	RuntimeBaseURL string
	Reason         string
}

type Step struct {
	ID                   StepID
	Name                 string
	Status               StepStatus
	Reason               string
	Method               string
	Command              string
	Notes                string
	Automatic            bool
	RequiresConfirmation bool
}

type InstallMethod struct {
	Name      string
	Command   string
	Notes     string
	Automatic bool
}

type Environment struct {
	Curl      system.CommandStatus
	Systemctl system.CommandStatus
	NPM       system.CommandStatus
	Bun       system.CommandStatus
	PNPM      system.CommandStatus
	Yarn      system.CommandStatus
	Brew      system.CommandStatus
	Python3   system.CommandStatus
	Pip       system.CommandStatus
	Pip3      system.CommandStatus
	UV        system.CommandStatus
	Pipx      system.CommandStatus
}

type Plan struct {
	Result                 doctor.Result
	Environment            Environment
	Recommendation         Recommendation
	Steps                  []Step
	OllamaInstallMethod    InstallMethod
	OllamaStartupMethod    InstallMethod
	OpenCodeInstallMethod  InstallMethod
	MemPalaceInstallMethod InstallMethod
}

func Build(ctx context.Context, cwd string) (Plan, error) {
	result, err := doctor.Collect(ctx, cwd)
	if err != nil {
		return Plan{}, err
	}

	environment := detectEnvironment()
	recommendation := buildRecommendation(result)
	ollamaInstall := planOllamaInstall(result, environment)
	ollamaStartup := planOllamaStartup(result, environment)
	openCodeInstall := planOpenCodeInstall(result, environment)
	memPalaceInstall := planMemPalaceInstall(result, environment)

	return Plan{
		Result:                 result,
		Environment:            environment,
		Recommendation:         recommendation,
		Steps:                  buildSteps(result, recommendation, ollamaInstall, ollamaStartup, openCodeInstall, memPalaceInstall),
		OllamaInstallMethod:    ollamaInstall,
		OllamaStartupMethod:    ollamaStartup,
		OpenCodeInstallMethod:  openCodeInstall,
		MemPalaceInstallMethod: memPalaceInstall,
	}, nil
}

func buildRecommendation(result doctor.Result) Recommendation {
	profile := detectProfile(result.CurrentDir, result.InGitRepo)
	companion := companions.DefaultForProfile(profile)
	modelRecommendation := models.RecommendCodingModel(result.RAMGB, result.GPU)
	shellProvider, shellEnabled := recommendShellProvider(result)

	return Recommendation{
		Profile:        profile,
		Model:          modelRecommendation.Recommended,
		FallbackModel:  modelRecommendation.Fallback,
		Agent:          "OpenCode",
		Memory:         "MemPalace",
		CompanionKey:   companion.Key,
		CompanionName:  companion.Name,
		ShellProvider:  shellProvider,
		ShellEnabled:   shellEnabled,
		RuntimeBaseURL: result.Ollama.BaseURL,
		Reason:         modelRecommendation.Reason,
	}
}

func detectEnvironment() Environment {
	return Environment{
		Curl:      system.DetectCommand("curl"),
		Systemctl: system.DetectCommand("systemctl"),
		NPM:       system.DetectCommand("npm"),
		Bun:       system.DetectCommand("bun"),
		PNPM:      system.DetectCommand("pnpm"),
		Yarn:      system.DetectCommand("yarn"),
		Brew:      system.DetectCommand("brew"),
		Python3:   system.DetectCommand("python3"),
		Pip:       system.DetectCommand("pip"),
		Pip3:      system.DetectCommand("pip3"),
		UV:        system.DetectCommand("uv"),
		Pipx:      system.DetectCommand("pipx"),
	}
}

func buildSteps(result doctor.Result, recommendation Recommendation, ollamaInstall, ollamaStartup, openCodeInstall, memPalaceInstall InstallMethod) []Step {
	steps := []Step{
		{
			ID:        StepCheckSystem,
			Name:      "Check system",
			Status:    StatusDone,
			Reason:    "Read-only machine diagnostics completed successfully.",
			Automatic: true,
		},
		globalConfigStep(result, recommendation),
		ollamaInstallStep(result, ollamaInstall),
		ollamaStartupStep(result, ollamaStartup),
		openCodeInstallStep(result, openCodeInstall),
		memPalaceInstallStep(result, memPalaceInstall),
		modelPlanStep(result, recommendation),
		{
			ID:        StepRegisterCompanions,
			Name:      "Install/register global companions",
			Status:    StatusSkipped,
			Reason:    "Companion registration remains scaffolded in the current Odin v2 implementation.",
			Automatic: false,
		},
		{
			ID:        StepShellIntegration,
			Name:      "Shell integration",
			Status:    StatusSkipped,
			Reason:    "Powerlevel10k and Starship integration are planned later and are not changed by setup yet.",
			Automatic: false,
		},
	}

	if result.InGitRepo {
		steps = append(steps, Step{
			ID:        StepProjectInit,
			Name:      "Project initialization",
			Status:    StatusSkipped,
			Reason:    "This directory is inside a Git repository. Odin will offer `odin init` after global setup is confirmed.",
			Command:   "odin init",
			Automatic: false,
		})
	}

	return steps
}

func globalConfigStep(result doctor.Result, recommendation Recommendation) Step {
	if result.GlobalConfig.Exists {
		return Step{
			ID:        StepGlobalConfig,
			Name:      "Create ~/.odin/config.toml",
			Status:    StatusDone,
			Reason:    "Global Odin config already exists.",
			Automatic: true,
		}
	}

	return Step{
		ID:                   StepGlobalConfig,
		Name:                 "Create ~/.odin/config.toml",
		Status:               StatusNeeded,
		Reason:               fmt.Sprintf("Odin global config is missing. Setup would create it with the recommended %s profile, %s model, and %s companion.", recommendation.Profile, recommendation.Model, recommendation.CompanionName),
		Method:               "Write Odin-managed global config",
		Automatic:            true,
		RequiresConfirmation: false,
	}
}

func ollamaInstallStep(result doctor.Result, method InstallMethod) Step {
	tool := result.Tool("ollama")
	if tool.Installed {
		return Step{
			ID:        StepInstallOllama,
			Name:      "Install Ollama",
			Status:    StatusDone,
			Reason:    fmt.Sprintf("Ollama is already installed at %s.", tool.Path),
			Automatic: false,
		}
	}

	status := StatusManual
	if result.OS.Name != "linux" {
		status = StatusUnsupported
	}
	if method.Automatic {
		status = StatusNeeded
	}

	return Step{
		ID:                   StepInstallOllama,
		Name:                 "Install Ollama",
		Status:               status,
		Reason:               "Ollama is missing.",
		Method:               method.Name,
		Command:              method.Command,
		Notes:                method.Notes,
		Automatic:            method.Automatic,
		RequiresConfirmation: true,
	}
}

func ollamaStartupStep(result doctor.Result, method InstallMethod) Step {
	if result.Ollama.APIAvailable {
		return Step{
			ID:        StepVerifyOllama,
			Name:      "Start or verify Ollama service/API",
			Status:    StatusDone,
			Reason:    fmt.Sprintf("Ollama API is already responding at %s.", result.Ollama.BaseURL),
			Automatic: false,
		}
	}

	if !result.Tool("ollama").Installed {
		status := StatusSkipped
		if result.OS.Name != "linux" {
			status = StatusUnsupported
		}
		return Step{
			ID:        StepVerifyOllama,
			Name:      "Start or verify Ollama service/API",
			Status:    status,
			Reason:    "Install Ollama first, then verify that its local API is responding.",
			Method:    method.Name,
			Command:   method.Command,
			Notes:     method.Notes,
			Automatic: false,
		}
	}

	status := StatusManual
	if result.OS.Name != "linux" {
		status = StatusUnsupported
	}
	if method.Automatic {
		status = StatusNeeded
	}

	return Step{
		ID:                   StepVerifyOllama,
		Name:                 "Start or verify Ollama service/API",
		Status:               status,
		Reason:               fmt.Sprintf("Ollama is installed but the local API at %s is not responding.", result.Ollama.BaseURL),
		Method:               method.Name,
		Command:              method.Command,
		Notes:                method.Notes,
		Automatic:            method.Automatic,
		RequiresConfirmation: true,
	}
}

func openCodeInstallStep(result doctor.Result, method InstallMethod) Step {
	tool := result.Tool("opencode")
	if tool.Installed {
		if strings.HasPrefix(tool.Details, "version check failed:") {
			return Step{
				ID:                   StepInstallOpenCode,
				Name:                 "Install OpenCode",
				Status:               StatusManual,
				Reason:               fmt.Sprintf("OpenCode was found at %s, but the CLI version check failed.", tool.Path),
				Notes:                tool.Details,
				Automatic:            false,
				RequiresConfirmation: true,
			}
		}
		if tool.Details != "" {
			return Step{
				ID:        StepInstallOpenCode,
				Name:      "Install OpenCode",
				Status:    StatusDone,
				Reason:    fmt.Sprintf("OpenCode is already installed at %s (%s).", tool.Path, tool.Details),
				Automatic: false,
			}
		}

		return Step{
			ID:        StepInstallOpenCode,
			Name:      "Install OpenCode",
			Status:    StatusDone,
			Reason:    fmt.Sprintf("OpenCode is already installed at %s.", tool.Path),
			Automatic: false,
		}
	}

	status := StatusManual
	if result.OS.Name != "linux" {
		status = StatusUnsupported
	}
	if method.Automatic {
		status = StatusNeeded
	}

	return Step{
		ID:                   StepInstallOpenCode,
		Name:                 "Install OpenCode",
		Status:               status,
		Reason:               "OpenCode is missing.",
		Method:               method.Name,
		Command:              method.Command,
		Notes:                method.Notes,
		Automatic:            method.Automatic,
		RequiresConfirmation: true,
	}
}

func memPalaceInstallStep(result doctor.Result, method InstallMethod) Step {
	tool := result.Tool("mempalace")
	if tool.Installed {
		if strings.HasPrefix(tool.Details, "mcp helper check failed:") {
			return Step{
				ID:                   StepInstallMemPalace,
				Name:                 "Install MemPalace",
				Status:               StatusManual,
				Reason:               fmt.Sprintf("MemPalace was found at %s, but the documented `mempalace mcp` helper did not complete cleanly.", tool.Path),
				Notes:                tool.Details,
				Automatic:            false,
				RequiresConfirmation: true,
			}
		}
		message := fmt.Sprintf("MemPalace is already installed at %s.", tool.Path)
		if tool.Details != "" {
			message = fmt.Sprintf("%s %s.", strings.TrimSuffix(message, "."), tool.Details)
		}
		return Step{
			ID:        StepInstallMemPalace,
			Name:      "Install MemPalace",
			Status:    StatusDone,
			Reason:    message,
			Automatic: false,
		}
	}

	status := StatusManual
	if result.OS.Name != "linux" {
		status = StatusUnsupported
	}
	if method.Automatic {
		status = StatusNeeded
	}

	return Step{
		ID:                   StepInstallMemPalace,
		Name:                 "Install MemPalace",
		Status:               status,
		Reason:               "MemPalace is missing.",
		Method:               method.Name,
		Command:              method.Command,
		Notes:                method.Notes,
		Automatic:            method.Automatic,
		RequiresConfirmation: true,
	}
}

func modelPlanStep(result doctor.Result, recommendation Recommendation) Step {
	if modelInstalled(result.Ollama.Models, recommendation.Model) {
		return Step{
			ID:        StepPullModel,
			Name:      "Pull recommended model",
			Status:    StatusDone,
			Reason:    fmt.Sprintf("Recommended model %q is already installed in Ollama.", recommendation.Model),
			Automatic: false,
		}
	}

	if result.Ollama.APIAvailable {
		return Step{
			ID:                   StepPullModel,
			Name:                 "Pull recommended model",
			Status:               StatusNeeded,
			Reason:               fmt.Sprintf("Recommended model %q is not installed locally.", recommendation.Model),
			Method:               "Use Odin model pull flow",
			Command:              fmt.Sprintf("odin model pull %s", recommendation.Model),
			Notes:                "Dry-run leaves model downloads to the explicit model pull flow.",
			Automatic:            true,
			RequiresConfirmation: true,
		}
	}

	return Step{
		ID:        StepPullModel,
		Name:      "Pull recommended model",
		Status:    StatusSkipped,
		Reason:    fmt.Sprintf("Wait until Ollama is installed and its API is responding at %s, then pull %q.", recommendation.RuntimeBaseURL, recommendation.Model),
		Command:   fmt.Sprintf("odin model pull %s", recommendation.Model),
		Automatic: false,
	}
}

func planOllamaInstall(result doctor.Result, environment Environment) InstallMethod {
	if result.OS.Name != "linux" {
		return InstallMethod{
			Name:      "Manual installation",
			Command:   "See the official Ollama download page for your platform",
			Notes:     "Automatic Odin setup is Linux-first for now.",
			Automatic: false,
		}
	}

	if environment.Curl.Installed {
		return InstallMethod{
			Name:      "Official Linux install script",
			Command:   "curl -fsSL https://ollama.com/install.sh | sh",
			Notes:     "Official docs also describe a manual tarball install and recommend a systemd service on Linux. Odin does not execute curl-pipe-shell installers automatically yet.",
			Automatic: false,
		}
	}

	return InstallMethod{
		Name:      "Manual Linux install",
		Command:   "Install curl first, then run the official Ollama install script or follow the manual tarball instructions",
		Notes:     "Ollama's Linux docs assume curl for the install script.",
		Automatic: false,
	}
}

func planOllamaStartup(result doctor.Result, environment Environment) InstallMethod {
	if result.OS.Name != "linux" {
		return InstallMethod{
			Name:      "Manual startup",
			Command:   "Start Ollama using the documented method for your platform",
			Automatic: false,
		}
	}

	if environment.Systemctl.Installed {
		return InstallMethod{
			Name:      "Recommended Linux systemd service",
			Command:   "sudo systemctl start ollama",
			Notes:     "Official Linux docs recommend enabling the ollama systemd service for startup.",
			Automatic: false,
		}
	}

	return InstallMethod{
		Name:      "Manual local serve",
		Command:   "ollama serve",
		Notes:     "Use this when systemd is unavailable or you prefer a foreground process.",
		Automatic: false,
	}
}

func planOpenCodeInstall(result doctor.Result, environment Environment) InstallMethod {
	if result.OS.Name != "linux" {
		return InstallMethod{
			Name:      "Manual installation",
			Command:   "See https://opencode.ai/docs/ for the current install methods for your platform",
			Notes:     "Automatic Odin setup is Linux-first for now.",
			Automatic: false,
		}
	}

	if environment.NPM.Installed {
		return InstallMethod{
			Name:      "Node.js npm global install",
			Command:   "npm install -g opencode-ai",
			Notes:     joinNonEmpty([]string{"Official OpenCode docs also support bun, pnpm, yarn, Homebrew, and an install script.", openCodeAlternatives(environment, "npm")}),
			Automatic: true,
		}
	}
	if environment.Bun.Installed {
		return InstallMethod{
			Name:      "Bun global install",
			Command:   "bun install -g opencode-ai",
			Notes:     joinNonEmpty([]string{"Official OpenCode docs also support npm, pnpm, yarn, Homebrew, and an install script.", openCodeAlternatives(environment, "bun")}),
			Automatic: true,
		}
	}
	if environment.PNPM.Installed {
		return InstallMethod{
			Name:      "pnpm global install",
			Command:   "pnpm install -g opencode-ai",
			Notes:     joinNonEmpty([]string{"Official OpenCode docs also support npm, bun, yarn, Homebrew, and an install script.", openCodeAlternatives(environment, "pnpm")}),
			Automatic: true,
		}
	}
	if environment.Yarn.Installed {
		return InstallMethod{
			Name:      "Yarn global install",
			Command:   "yarn global add opencode-ai",
			Notes:     joinNonEmpty([]string{"Official OpenCode docs also support npm, bun, pnpm, Homebrew, and an install script.", openCodeAlternatives(environment, "yarn")}),
			Automatic: true,
		}
	}
	if environment.Brew.Installed {
		return InstallMethod{
			Name:      "Homebrew install",
			Command:   "brew install anomalyco/tap/opencode",
			Notes:     "OpenCode docs recommend the anomalyco tap for the most up-to-date releases.",
			Automatic: true,
		}
	}

	fallback := InstallMethod{
		Name:      "Manual installation",
		Command:   "See https://opencode.ai/docs/ for the official install script or package manager methods",
		Notes:     "Odin did not find npm, bun, pnpm, yarn, or Homebrew on this Linux machine.",
		Automatic: false,
	}
	if environment.Curl.Installed {
		fallback.Notes = "Official docs list an install script and multiple package manager methods. Odin is leaving the install script as a manual step for now."
	}

	return fallback
}

func planMemPalaceInstall(result doctor.Result, environment Environment) InstallMethod {
	if result.OS.Name != "linux" {
		return InstallMethod{
			Name:      "Manual installation",
			Command:   "pip install mempalace",
			Notes:     "MemPalace docs install from PyPI with pip. Automatic Odin setup is Linux-first for now.",
			Automatic: false,
		}
	}

	if environment.Python3.Installed && environment.Pip.Installed {
		return InstallMethod{
			Name:      "PyPI install with pip",
			Command:   "pip install mempalace",
			Notes:     "Official MemPalace docs list pip as the primary install path and require Python 3.9+.",
			Automatic: true,
		}
	}

	notes := []string{"Official MemPalace docs currently show `pip install mempalace`."}
	switch {
	case !environment.Python3.Installed:
		notes = append(notes, "Python 3.9+ is not currently available on PATH.")
	case environment.Pip3.Installed:
		notes = append(notes, "pip3 is available, but Odin is keeping installation manual because the docs currently show pip specifically.")
	case environment.UV.Installed || environment.Pipx.Installed:
		notes = append(notes, "uv/pipx were detected, but Odin is not planning them automatically because the current MemPalace docs do not document those install paths.")
	default:
		notes = append(notes, "pip is not currently available on PATH.")
	}

	return InstallMethod{
		Name:      "Manual Python setup",
		Command:   "Install Python 3.9+ with pip, then run: pip install mempalace",
		Notes:     strings.Join(notes, " "),
		Automatic: false,
	}
}

func openCodeAlternatives(environment Environment, preferred string) string {
	options := []string{}
	if preferred != "npm" && environment.NPM.Installed {
		options = append(options, "npm install -g opencode-ai")
	}
	if preferred != "bun" && environment.Bun.Installed {
		options = append(options, "bun install -g opencode-ai")
	}
	if preferred != "pnpm" && environment.PNPM.Installed {
		options = append(options, "pnpm install -g opencode-ai")
	}
	if preferred != "yarn" && environment.Yarn.Installed {
		options = append(options, "yarn global add opencode-ai")
	}
	if preferred != "brew" && environment.Brew.Installed {
		options = append(options, "brew install anomalyco/tap/opencode")
	}

	if len(options) == 0 {
		return ""
	}

	return "Available alternatives: " + strings.Join(options, "; ")
}

func recommendShellProvider(result doctor.Result) (string, bool) {
	if result.Powerlevel10kConfigured {
		return "p10k", true
	}
	if result.Tool("starship").Installed {
		return "starship", true
	}
	return "p10k", false
}

func detectProfile(cwd string, inGitRepo bool) string {
	if inGitRepo || hasProjectMarkers(cwd) {
		return "developer"
	}
	return "beginner"
}

func hasProjectMarkers(dir string) bool {
	markers := []string{
		"go.mod",
		"package.json",
		"pyproject.toml",
		"Cargo.toml",
		".git",
		"Makefile",
	}

	for _, marker := range markers {
		if _, err := os.Stat(filepath.Join(dir, marker)); err == nil {
			return true
		}
	}

	return false
}

func modelInstalled(installed []ollamaplugin.Model, name string) bool {
	for _, model := range installed {
		if model.Name == name {
			return true
		}
	}

	return false
}

func joinNonEmpty(parts []string) string {
	filtered := []string{}
	for _, part := range parts {
		if strings.TrimSpace(part) != "" {
			filtered = append(filtered, part)
		}
	}
	return strings.Join(filtered, " ")
}
