package shell

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/danielsampar12/odin/internal/plugins"
	"github.com/danielsampar12/odin/internal/system"
)

type PromptStatus struct {
	Starship                plugins.Status
	Powerlevel10kConfigured bool
	Powerlevel10kSource     string
}

func Detect(home string) PromptStatus {
	starship := system.DetectCommand("starship")
	status := PromptStatus{
		Starship: plugins.Status{
			Name:      "Starship",
			Command:   "starship",
			Installed: starship.Installed,
			Path:      starship.Path,
		},
	}

	if home == "" {
		return status
	}

	for _, candidate := range []string{
		filepath.Join(home, ".p10k.zsh"),
		filepath.Join(home, ".zshrc"),
		filepath.Join(home, ".zprofile"),
	} {
		matched, err := fileLooksLikePowerlevel10k(candidate)
		if err != nil || !matched {
			continue
		}

		status.Powerlevel10kConfigured = true
		status.Powerlevel10kSource = candidate
		return status
	}

	return status
}

func fileLooksLikePowerlevel10k(path string) (bool, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}

	content := string(body)
	return strings.Contains(content, "powerlevel10k") ||
		strings.Contains(content, ".p10k.zsh") ||
		strings.Contains(content, "POWERLEVEL9K"), nil
}
