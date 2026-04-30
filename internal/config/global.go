package config

import (
	"fmt"
	"os"
	"strings"
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

func ReadGlobalProfile(path string) (string, error) {
	return readQuotedConfigValue(path, "", "profile")
}

func ReadGlobalRuntimeProvider(path string) (string, error) {
	return readQuotedConfigValue(path, "runtime", "provider")
}

func ReadGlobalRuntimeBaseURL(path string) (string, error) {
	return readQuotedConfigValue(path, "runtime", "base_url")
}

func ReadGlobalMemoryProvider(path string) (string, error) {
	return readQuotedConfigValue(path, "memory", "provider")
}

func ReadGlobalModelDefault(path string) (string, error) {
	return readQuotedConfigValue(path, "model", "default")
}

func ReadGlobalCompanionDefault(path string) (string, error) {
	return readQuotedConfigValue(path, "companion", "default")
}

func ResolveGlobalRuntimeBaseURL(path, fallback string) string {
	baseURL, err := ReadGlobalRuntimeBaseURL(path)
	if err != nil || baseURL == "" {
		return fallback
	}

	return baseURL
}

func quotedValue(line string) string {
	start := strings.Index(line, `"`)
	end := strings.LastIndex(line, `"`)
	if start == -1 || end <= start {
		return ""
	}
	return line[start+1 : end]
}

func readQuotedConfigValue(path, section, key string) (string, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	currentSection := ""
	for _, rawLine := range strings.Split(string(body), "\n") {
		line := strings.TrimSpace(rawLine)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			currentSection = strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(line, "["), "]"))
			continue
		}

		if section == "" {
			if currentSection != "" {
				continue
			}
		} else if currentSection != section {
			continue
		}

		if !strings.HasPrefix(line, key) {
			continue
		}

		if value := quotedValue(line); value != "" {
			return value, nil
		}
	}

	return "", nil
}
