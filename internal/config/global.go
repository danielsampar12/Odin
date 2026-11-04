package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

type GlobalSettings struct {
	Profile          string
	RuntimeProvider  string
	RuntimeBaseURL   string
	AgentProvider    string
	MemoryProvider   string
	ShellProvider    string
	ShellEnabled     bool
	ModelDefault     string
	ModelFallback    string
	CompanionDefault string
}

func DefaultGlobalConfig(settings GlobalSettings) string {
	return fmt.Sprintf(`profile = %q

[runtime]
provider = %q
base_url = %q

[agent]
provider = %q

[memory]
provider = %q

[shell]
provider = %q
enabled = %t

[model]
default = %q
fallback = %q

[companion]
default = %q
`, settings.Profile, settings.RuntimeProvider, settings.RuntimeBaseURL, settings.AgentProvider, settings.MemoryProvider, settings.ShellProvider, settings.ShellEnabled, settings.ModelDefault, settings.ModelFallback, settings.CompanionDefault)
}

// File is one Odin config file: ~/.odin/config.toml (machine defaults) or .odin/config.toml (shared
// project intent). Keys a file does not set are empty.
type File struct {
	Name    string `toml:"name"`
	Profile string `toml:"profile"`
	Runtime struct {
		Provider string `toml:"provider"`
		BaseURL  string `toml:"base_url"`
	} `toml:"runtime"`
	Agent struct {
		Provider string `toml:"provider"`
	} `toml:"agent"`
	Memory struct {
		Provider string `toml:"provider"`
		Hall     string `toml:"hall"`
	} `toml:"memory"`
	Shell struct {
		Provider string `toml:"provider"`
		Enabled  bool   `toml:"enabled"`
	} `toml:"shell"`
	Model struct {
		Default  string `toml:"default"`
		Fallback string `toml:"fallback"`
	} `toml:"model"`
	Companion struct {
		Default string `toml:"default"`
	} `toml:"companion"`
}

// Load reads an Odin config file. A missing file is an empty File; invalid TOML is an error naming the file.
func Load(path string) (File, error) {
	var file File
	if _, err := toml.DecodeFile(path, &file); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return File{}, nil
		}
		return File{}, fmt.Errorf("read %s: %w", path, err)
	}
	return file, nil
}

// ResolveGlobalRuntimeBaseURL returns the Ollama URL from the global config, or fallback when it is unset or unreadable.
func ResolveGlobalRuntimeBaseURL(path, fallback string) string {
	file, err := Load(path)
	if err != nil || file.Runtime.BaseURL == "" {
		return fallback
	}
	return file.Runtime.BaseURL
}

// WriteFileIfMissing writes content to path unless the file already exists, and reports whether it wrote.
func WriteFileIfMissing(path, content string, perm os.FileMode) (bool, error) {
	if _, err := os.Stat(path); err == nil {
		return false, nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return false, err
	}

	return true, os.WriteFile(path, []byte(content), perm)
}
