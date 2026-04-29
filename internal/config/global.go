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
	body, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	for _, line := range strings.Split(string(body), "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "profile") {
			continue
		}

		if value := quotedValue(line); value != "" {
			return value, nil
		}
	}

	return "", nil
}

func quotedValue(line string) string {
	start := strings.Index(line, `"`)
	end := strings.LastIndex(line, `"`)
	if start == -1 || end <= start {
		return ""
	}
	return line[start+1 : end]
}
