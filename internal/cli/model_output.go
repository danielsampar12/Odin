package cli

import (
	"fmt"
	"strings"

	ollamaplugin "github.com/danielsampar12/odin/internal/plugins/ollama"
)

func ollamaAPIStatusLine(status ollamaplugin.APIStatus) string {
	if status.APIAvailable {
		return fmt.Sprintf("reachable (%s)", status.BaseURL)
	}
	if status.Error != "" {
		return fmt.Sprintf("unavailable (%s: %s)", status.BaseURL, status.Error)
	}
	return fmt.Sprintf("unavailable (%s)", status.BaseURL)
}

func summarizeModelNames(models []ollamaplugin.Model, limit int) string {
	if len(models) == 0 {
		return "none installed"
	}

	names := make([]string, 0, len(models))
	for _, model := range models {
		names = append(names, model.Name)
	}

	if len(names) <= limit {
		return strings.Join(names, ", ")
	}

	return fmt.Sprintf("%s, +%d more", strings.Join(names[:limit], ", "), len(names)-limit)
}

func formatModelEntry(model ollamaplugin.Model) string {
	parts := make([]string, 0, 2)
	if model.Details.ParameterSize != "" {
		parts = append(parts, model.Details.ParameterSize)
	}
	if model.Details.QuantizationLevel != "" {
		parts = append(parts, model.Details.QuantizationLevel)
	}
	if len(parts) == 0 {
		return model.Name
	}
	return fmt.Sprintf("%s (%s)", model.Name, strings.Join(parts, ", "))
}
