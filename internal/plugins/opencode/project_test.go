package opencode

import (
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/danielsampar12/odin/internal/config"
	"github.com/danielsampar12/odin/internal/models"
)

func TestWriteGeneratedConfig(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	projectDir := t.TempDir()
	if err := os.MkdirAll(config.ProjectDir(projectDir), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(config.GlobalDir(), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(config.GlobalConfigPath(), []byte(`profile = "developer"

[runtime]
provider = "ollama"
base_url = "http://localhost:11434"

[model]
default = "global-model"

[companion]
default = "freya"
`), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(config.ProjectConfigPath(projectDir), []byte(`name = "demo"

[runtime]
provider = "ollama"

[model]
default = "project-model"

[companion]
default = "baldur"
`), 0o644); err != nil {
		t.Fatal(err)
	}

	scaffold, err := ResolveProjectScaffold(projectDir)
	if err != nil {
		t.Fatal(err)
	}
	if scaffold.Model != "global-model" || scaffold.Companion != "baldur" {
		t.Fatalf("model, companion = %q, %q; want the global model (D4) and the project companion", scaffold.Model, scaffold.Companion)
	}
	if len(scaffold.Warnings) != 1 || !strings.Contains(scaffold.Warnings[0], "[model] default") {
		t.Fatalf("warnings = %q, want one about the ignored project model", scaffold.Warnings)
	}

	result, err := WriteGeneratedConfig(projectDir, GenerateOptions{Context: 16384})
	if err != nil {
		t.Fatalf("WriteGeneratedConfig error = %v", err)
	}
	if !result.Written {
		t.Fatalf("expected config to be written")
	}

	body, err := os.ReadFile(config.ProjectGeneratedOpenCodeConfigPath(projectDir))
	if err != nil {
		t.Fatal(err)
	}
	content := string(body)
	if !strings.Contains(content, ManagedMarker) {
		t.Fatalf("generated config is missing managed marker")
	}
	if !strings.Contains(content, `"model": "ollama/global-model:latest-ctx16k"`) {
		t.Fatalf("generated config is missing the context model reference: %s", content)
	}
	if !strings.Contains(content, `"context": 16384`) || !strings.Contains(content, `"output": 4096`) {
		t.Fatalf("generated config is missing the model context limit: %s", content)
	}
	if !GeneratedConfigUses(projectDir, "ollama/global-model:latest-ctx16k") || GeneratedConfigUses(projectDir, "ollama/global-model:latest-ctx32k") {
		t.Fatalf("GeneratedConfigUses does not match the generated model")
	}
	if !strings.Contains(content, `"baseURL": "http://localhost:11434/v1"`) {
		t.Fatalf("generated config is missing Ollama v1 base URL: %s", content)
	}
	if !strings.Contains(content, `"instructions": [`) || !strings.Contains(content, `".odin/rules.md"`) {
		t.Fatalf("generated config is missing rules instruction reference: %s", content)
	}
}

func TestWriteGeneratedConfigRefusesUnmanagedFile(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	projectDir := t.TempDir()
	if err := os.MkdirAll(config.ProjectGeneratedDir(projectDir), 0o755); err != nil {
		t.Fatal(err)
	}

	path := config.ProjectGeneratedOpenCodeConfigPath(projectDir)
	if err := os.WriteFile(path, []byte(`{"custom":true}`), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := WriteGeneratedConfig(projectDir, GenerateOptions{Context: 16384})
	if !errors.Is(err, ErrUnmanagedGeneratedConfig) {
		t.Fatalf("error = %v, want ErrUnmanagedGeneratedConfig", err)
	}

	result, err := WriteGeneratedConfig(projectDir, GenerateOptions{Force: true, Context: 16384})
	if err != nil {
		t.Fatalf("force WriteGeneratedConfig error = %v", err)
	}
	if !result.Written {
		t.Fatalf("expected forced regeneration to write file")
	}
}

func TestWriteGeneratedConfigWithMemory(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	projectDir := t.TempDir()
	if err := os.MkdirAll(config.ProjectDir(projectDir), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(config.ProjectConfigPath(projectDir), []byte(`name = "demo"

[runtime]
provider = "ollama"

[memory]
provider = "mempalace"
hall = "demo-hall"

[model]
default = "project-model"
`), 0o644); err != nil {
		t.Fatal(err)
	}

	result, err := WriteGeneratedConfig(projectDir, GenerateOptions{WithMemory: true, Context: 16384})
	if err != nil {
		t.Fatalf("WriteGeneratedConfig error = %v", err)
	}
	if !result.Written {
		t.Fatalf("expected config to be written")
	}

	body, err := os.ReadFile(config.ProjectGeneratedOpenCodeConfigPath(projectDir))
	if err != nil {
		t.Fatal(err)
	}
	content := string(body)
	if !strings.Contains(content, `// MemPalace MCP: enabled`) {
		t.Fatalf("generated config is missing MemPalace marker: %s", content)
	}
	if !strings.Contains(content, `"mempalace"`) || !strings.Contains(content, `mempalace.mcp_server`) {
		t.Fatalf("generated config is missing MemPalace MCP wiring: %s", content)
	}
}

func TestGeneratedConfigDrift(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	projectDir := t.TempDir()
	path := config.ProjectGeneratedOpenCodeConfigPath(projectDir)
	state := func(context int) ConfigState {
		t.Helper()
		got, err := GeneratedConfigState(projectDir, GenerateOptions{Context: context})
		if err != nil {
			t.Fatal(err)
		}
		return got
	}

	if got := state(16384); got != ConfigMissing {
		t.Fatalf("state before generate = %v, want missing", got)
	}
	if _, err := WriteGeneratedConfig(projectDir, GenerateOptions{Context: 16384}); err != nil {
		t.Fatal(err)
	}
	if got := state(16384); got != ConfigUpToDate {
		t.Fatalf("state after generate = %v, want up to date", got)
	}
	if got := state(32768); got != ConfigOutdated {
		t.Fatalf("state for a new context size = %v, want out of date", got)
	}
	result, err := WriteGeneratedConfig(projectDir, GenerateOptions{Context: 32768})
	if err != nil || !result.Written || result.State != ConfigOutdated {
		t.Fatalf("regenerate out-of-date config: result %+v, err %v; want it rewritten", result, err)
	}

	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	edited := strings.Replace(string(body), `"output": 8192`, `"output": 4096`, 1)
	if err := os.WriteFile(path, []byte(edited), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := state(32768); got != ConfigEdited {
		t.Fatalf("state after a hand edit = %v, want edited by hand", got)
	}
	if _, err := WriteGeneratedConfig(projectDir, GenerateOptions{Context: 32768}); !errors.Is(err, ErrEditedGeneratedConfig) {
		t.Fatalf("generate over a hand edit: err = %v, want ErrEditedGeneratedConfig", err)
	}
	if after, _ := os.ReadFile(path); string(after) != edited {
		t.Fatalf("generate changed a hand-edited config without --force")
	}
	if _, err := WriteGeneratedConfig(projectDir, GenerateOptions{Context: 32768, Force: true}); err != nil {
		t.Fatalf("generate --force: %v", err)
	}

	// A CRLF checkout of an untouched file is not a hand edit.
	if got := classify(strings.ReplaceAll(string(body), "\n", "\r\n"), string(body)); got != ConfigUpToDate {
		t.Fatalf("classify(CRLF copy) = %v, want up to date", got)
	}
	// A line added above the header moves the marker off line 1: that is a hand edit, not a v0.2 file.
	if got := classify("// my note\n"+string(body), string(body)); got != ConfigEdited {
		t.Fatalf("classify(line added above the header) = %v, want edited by hand", got)
	}

	// Odin v0.2 wrote the marker without a hash.
	legacy := "// " + ManagedMarker + ".\n{}\n"
	if err := os.WriteFile(path, []byte(legacy), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := state(32768); got != ConfigOutdated {
		t.Fatalf("state of a v0.2 config = %v, want out of date", got)
	}
}

func TestEnsureV1Endpoint(t *testing.T) {
	testCases := []struct {
		input string
		want  string
	}{
		{input: "http://localhost:11434", want: "http://localhost:11434/v1"},
		{input: "http://localhost:11434/", want: "http://localhost:11434/v1"},
		{input: "http://localhost:11434/v1", want: "http://localhost:11434/v1"},
	}

	for _, tc := range testCases {
		if got := ensureV1Endpoint(tc.input); got != tc.want {
			t.Fatalf("ensureV1Endpoint(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestGeneratedConfigHasCompanionAgents(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	projectDir := t.TempDir()
	if err := os.MkdirAll(config.ProjectDir(projectDir), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(config.ProjectConfigPath(projectDir), []byte("[companion]\ndefault = \"tyr\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := WriteGeneratedConfig(projectDir, GenerateOptions{Context: 32768}); err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(config.ProjectGeneratedOpenCodeConfigPath(projectDir))
	if err != nil {
		t.Fatal(err)
	}
	var generated struct {
		DefaultAgent string `json:"default_agent"`
		Agent        map[string]struct {
			Mode       string         `json:"mode"`
			Prompt     string         `json:"prompt"`
			Permission map[string]any `json:"permission"`
		} `json:"agent"`
	}
	if err := json.Unmarshal(stripJSONC(body), &generated); err != nil {
		t.Fatal(err)
	}
	if generated.DefaultAgent != "tyr" {
		t.Fatalf("default_agent = %q, want the project's companion", generated.DefaultAgent)
	}
	baldur, tyr := generated.Agent["baldur"], generated.Agent["tyr"]
	if baldur.Mode != "primary" || baldur.Prompt == "" || baldur.Permission["task"] != "deny" || baldur.Permission["edit"] != nil {
		t.Fatalf("baldur agent = %+v", baldur)
	}
	if tyr.Permission["edit"] != "deny" || tyr.Permission["bash"] != "deny" || tyr.Permission["task"] != "deny" {
		t.Fatalf("tyr must not edit files, run commands or start subagents: %+v", tyr.Permission)
	}

	if err := os.WriteFile(config.ProjectConfigPath(projectDir), []byte("[companion]\ndefault = \"freya\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	scaffold, err := ResolveProjectScaffold(projectDir)
	if err != nil {
		t.Fatal(err)
	}
	if scaffold.Companion != "baldur" || len(scaffold.Warnings) != 1 {
		t.Fatalf("a v1 companion should fall back to baldur with a warning: %q, %q", scaffold.Companion, scaffold.Warnings)
	}
}

func TestModelEntrySampling(t *testing.T) {
	plain := modelEntry("m", 32768, nil)
	if plain["temperature"] != nil || plain["options"] != nil || plain["variants"] != nil {
		t.Fatalf("a model without catalog sampling should keep Ollama's settings: %v", plain)
	}
	entry := modelEntry("m", 65536, &models.Sampling{Temperature: 0.6, NoThinking: true})
	if entry["temperature"] != true {
		t.Fatalf("OpenCode only sends the agent's temperature when the model allows it: %v", entry)
	}
	if entry["options"].(map[string]any)["reasoningEffort"] != "none" {
		t.Fatalf("thinking should be off by default: %v", entry)
	}
	if think := entry["variants"].(map[string]any)["think"].(map[string]any); think["reasoningEffort"] != "high" || think["max_tokens"] != 16384 {
		t.Fatalf("the think variant should turn thinking back on with room to think: %v", entry)
	}
	if limit := entry["limit"].(map[string]int); limit["output"] != 8192 {
		t.Fatalf("thinking off should reserve 8192 output tokens at 64k: %v", limit)
	}
	if limit := modelEntry("m", 16384, &models.Sampling{NoThinking: true})["limit"].(map[string]int); limit["output"] != 4096 {
		t.Fatalf("a small context should keep a quarter for output: %v", limit)
	}
}
