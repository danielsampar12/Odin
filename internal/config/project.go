package config

import "fmt"

type ProjectSettings struct {
	Name             string
	AgentProvider    string
	RuntimeProvider  string
	MemoryProvider   string
	MemoryHall       string
	CompanionDefault string
}

// DefaultProjectConfig is .odin/config.toml: shared intent that is safe to commit. Machine-specific
// values (model, Ollama URL) belong in ~/.odin/config.toml.
func DefaultProjectConfig(settings ProjectSettings) string {
	return fmt.Sprintf(`name = %q

[agent]
provider = %q

[runtime]
provider = %q

[memory]
provider = %q
hall = %q

[companion]
default = %q
`, settings.Name, settings.AgentProvider, settings.RuntimeProvider, settings.MemoryProvider, settings.MemoryHall, settings.CompanionDefault)
}

// DefaultRules is .odin/rules.md. OpenCode puts it in every session's prompt, so it starts nearly empty.
func DefaultRules() string {
	return `# Project rules

<!-- OpenCode loads this file into every session in this repository. Add the rules the agent should
follow here: how to build and test, code style, what to avoid. -->
`
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

This file is intentionally small and inspectable.
`, projectName, companionName)
}
