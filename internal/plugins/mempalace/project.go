package mempalace

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/danielsampar12/odin/internal/config"
)

const ExpectedProvider = "mempalace"

var defaultMCPCommand = []string{"python", "-m", "mempalace.mcp_server"}

type ProjectStatus struct {
	GlobalProvider          string
	ProjectProvider         string
	ResolvedProvider        string
	ProjectConfigPath       string
	ProjectConfigExists     bool
	Hall                    string
	HallDerived             bool
	PalaceConfigPath        string
	PalaceConfigExists      bool
	PalaceIdentityPath      string
	PalaceIdentityExists    bool
	PalacePath              string
	OpenCodeConfigPath      string
	OpenCodeConfigExists    bool
	OpenCodeMCPConfigured   bool
	OpenCodeMCPEnabled      bool
	OpenCodeMCPCommandKnown bool
}

type palaceConfig struct {
	PalacePath string `json:"palace_path"`
}

func ResolveProjectStatus(cwd string) (ProjectStatus, error) {
	globalConfigPath := config.GlobalConfigPath()
	projectConfigPath := config.ProjectConfigPath(cwd)
	globalProvider, _ := config.ReadGlobalMemoryProvider(globalConfigPath)
	projectProvider, _ := config.ReadProjectMemoryProvider(projectConfigPath)
	hall, _ := config.ReadProjectMemoryHall(projectConfigPath)
	hallDerived := false
	if hall == "" {
		hall = DeriveHall(cwd)
		hallDerived = true
	}

	resolvedProvider := ExpectedProvider
	if globalProvider != "" {
		resolvedProvider = globalProvider
	}
	if projectProvider != "" {
		resolvedProvider = projectProvider
	}

	palaceConfigPath := GlobalConfigPath()
	palacePath, palaceConfigExists, err := resolvePalacePath(palaceConfigPath)
	if err != nil {
		return ProjectStatus{}, err
	}

	openCodeConfigPath := config.ProjectGeneratedOpenCodeConfigPath(cwd)
	mcpStatus, err := DetectOpenCodeMCP(openCodeConfigPath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return ProjectStatus{}, err
	}

	return ProjectStatus{
		GlobalProvider:          globalProvider,
		ProjectProvider:         projectProvider,
		ResolvedProvider:        resolvedProvider,
		ProjectConfigPath:       projectConfigPath,
		ProjectConfigExists:     fileExists(projectConfigPath),
		Hall:                    hall,
		HallDerived:             hallDerived,
		PalaceConfigPath:        palaceConfigPath,
		PalaceConfigExists:      palaceConfigExists,
		PalaceIdentityPath:      IdentityPath(),
		PalaceIdentityExists:    fileExists(IdentityPath()),
		PalacePath:              palacePath,
		OpenCodeConfigPath:      openCodeConfigPath,
		OpenCodeConfigExists:    fileExists(openCodeConfigPath),
		OpenCodeMCPConfigured:   mcpStatus.Configured,
		OpenCodeMCPEnabled:      mcpStatus.Enabled,
		OpenCodeMCPCommandKnown: mcpStatus.CommandKnown,
	}, nil
}

type OpenCodeMCPStatus struct {
	Configured   bool
	Enabled      bool
	CommandKnown bool
}

func DetectOpenCodeMCP(path string) (OpenCodeMCPStatus, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return OpenCodeMCPStatus{}, err
	}

	content := string(body)
	configured := strings.Contains(content, `"mempalace"`) && strings.Contains(content, "mempalace.mcp_server")
	enabled := configured && strings.Contains(content, `"enabled": true`)

	return OpenCodeMCPStatus{
		Configured:   configured,
		Enabled:      enabled,
		CommandKnown: configured && strings.Contains(content, `"python"`) && strings.Contains(content, "mempalace.mcp_server"),
	}, nil
}

func MCPCommand() []string {
	return append([]string{}, defaultMCPCommand...)
}

func GlobalConfigPath() string {
	return filepath.Join(GlobalDir(), "config.json")
}

func GlobalDir() string {
	home := config.HomeDir()
	if home == "" {
		return ".mempalace"
	}

	return filepath.Join(home, ".mempalace")
}

func IdentityPath() string {
	return filepath.Join(GlobalDir(), "identity.txt")
}

func DefaultPalacePath() string {
	return filepath.Join(GlobalDir(), "palace")
}

func DeriveHall(cwd string) string {
	base := filepath.Base(cwd)
	if base == "." || base == string(filepath.Separator) || base == "" {
		return "current-project"
	}

	return base
}

func resolvePalacePath(path string) (string, bool, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return DefaultPalacePath(), false, nil
		}
		return "", false, err
	}

	var parsed palaceConfig
	if err := json.Unmarshal(body, &parsed); err != nil {
		return DefaultPalacePath(), true, nil
	}

	palacePath := strings.TrimSpace(parsed.PalacePath)
	if palacePath == "" {
		palacePath = DefaultPalacePath()
	}

	return palacePath, true, nil
}

func fileExists(path string) bool {
	if _, err := os.Stat(path); err == nil {
		return true
	}

	return false
}
