package cli

import (
	"testing"

	ollamaplugin "github.com/danielsampar12/odin/internal/plugins/ollama"
)

func TestOllamaModelInstalledAcceptsImplicitLatestTag(t *testing.T) {
	models := []ollamaplugin.Model{{Name: "llama3.1:latest"}, {Name: "qwen3.5:9b-q4_K_M"}}
	for _, name := range []string{"llama3.1", "llama3.1:latest", "qwen3.5:9b-q4_K_M"} {
		if !ollamaModelInstalled(models, name) {
			t.Fatalf("ollamaModelInstalled(%q) = false, want true", name)
		}
	}
	if ollamaModelInstalled(models, "qwen3.5:9b") {
		t.Fatalf("ollamaModelInstalled matched a different tag")
	}
}
