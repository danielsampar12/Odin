package config

import "fmt"

type ProjectSettings struct {
	Name             string
	AgentProvider    string
	RuntimeProvider  string
	MemoryProvider   string
	MemoryHall       string
	ModelDefault     string
	CompanionDefault string
}

func DefaultProjectConfig(settings ProjectSettings) string {
	return fmt.Sprintf(`name = %q

[agent]
provider = %q

[runtime]
provider = %q

[memory]
provider = %q
hall = %q

[model]
default = %q

[companion]
default = %q
`, settings.Name, settings.AgentProvider, settings.RuntimeProvider, settings.MemoryProvider, settings.MemoryHall, settings.ModelDefault, settings.CompanionDefault)
}

func DefaultRules(projectName, companionName string) string {
	return fmt.Sprintf(`# Odin Project Rules

This project was initialized by Odin v2 for %s.

- Keep generated configs inspectable.
- Prefer local-first and privacy-first defaults.
- Treat OpenCode as the coding agent, Ollama as the runtime, and MemPalace as the primary memory provider.
- Use %s as the default Odin companion unless the project decides otherwise.
- Keep .odin/generated/ as Odin-managed generated output.
`, projectName, companionName)
}

func DefaultAgents(projectName, companionName string) string {
	return fmt.Sprintf(`# AGENTS

Odin project marker: odin-v2

This repository is scaffolded for Odin v2 and is intended to run local-first tooling.

Project: %s
Selected Odin companion: %s

Project instructions:
- Read .odin/rules.md alongside this file.
- Keep Odin-generated files inspectable and project-local.
- Prefer local runtimes and local models where possible.

Intended stack:
- Odin manages the local-first AI workstation experience for this project.
- OpenCode is the coding agent.
- Ollama is the local model runtime.
- MemPalace is the primary memory provider.

This file is intentionally small and inspectable.
`, projectName, companionName)
}

func ReadProjectRuntimeProvider(path string) (string, error) {
	return readQuotedConfigValue(path, "runtime", "provider")
}

func ReadProjectModelDefault(path string) (string, error) {
	return readQuotedConfigValue(path, "model", "default")
}

func ReadProjectCompanionDefault(path string) (string, error) {
	return readQuotedConfigValue(path, "companion", "default")
}
