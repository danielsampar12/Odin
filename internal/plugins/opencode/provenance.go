package opencode

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/danielsampar12/odin/internal/config"
)

// Source is one OpenCode config that applies when Odin launches OpenCode in a project.
type Source struct {
	Name  string   // display name: a project-relative path, a ~/ path or an environment variable
	Keys  []string // which of trackedKeys it sets
	Model string   // its "model" value, if it sets one
	Err   error    // the file exists but could not be parsed
	Odin  bool     // Odin's generated config
}

// trackedKeys are the settings worth tracing back to a source.
var trackedKeys = []string{"model", "provider", "agent", "default_agent", "permission", "instructions", "mcp"}

// Sources lists the OpenCode configs that exist for a launch from cwd, in the order OpenCode 1.19 merges
// them (later ones override earlier ones; arrays such as instructions are concatenated). Odin's generated
// config fills the OPENCODE_CONFIG slot because Odin sets that variable when it launches OpenCode.
func Sources(cwd string) []Source {
	var sources []Source
	add := func(name, path string) {
		if body, err := os.ReadFile(path); err == nil {
			sources = append(sources, parseSource(name, body))
		}
	}

	globalDir := GlobalConfigDir()
	for _, name := range []string{"config.json", "opencode.json", "opencode.jsonc"} {
		add(displayHome(filepath.Join(globalDir, name)), filepath.Join(globalDir, name))
	}
	add("OPENCODE_CONFIG="+filepath.ToSlash(RelativeConfigPath()), config.ProjectGeneratedOpenCodeConfigPath(cwd))
	if n := len(sources); n > 0 && strings.HasPrefix(sources[n-1].Name, "OPENCODE_CONFIG=") {
		sources[n-1].Odin = true
	}
	for _, name := range []string{"opencode.json", "opencode.jsonc", ".opencode/opencode.json", ".opencode/opencode.jsonc"} {
		add(name, filepath.Join(cwd, filepath.FromSlash(name)))
	}
	if content := os.Getenv("OPENCODE_CONFIG_CONTENT"); content != "" {
		sources = append(sources, parseSource("OPENCODE_CONFIG_CONTENT", []byte(content)))
	}
	return sources
}

// GlobalConfigDir is where OpenCode keeps its user config: $XDG_CONFIG_HOME/opencode, else ~/.config/opencode.
func GlobalConfigDir() string {
	if dir := os.Getenv("XDG_CONFIG_HOME"); dir != "" {
		return filepath.Join(dir, "opencode")
	}
	return filepath.Join(config.HomeDir(), ".config", "opencode")
}

// ModelOverrides returns the sources after Odin's generated config that set "model", which wins over Odin's.
func ModelOverrides(sources []Source) []Source {
	var overrides []Source
	afterOdin := false
	for _, source := range sources {
		if afterOdin && source.Model != "" {
			overrides = append(overrides, source)
		}
		afterOdin = afterOdin || source.Odin
	}
	return overrides
}

func parseSource(name string, body []byte) Source {
	source := Source{Name: name}
	var values map[string]json.RawMessage
	if err := json.Unmarshal(stripJSONC(body), &values); err != nil {
		source.Err = fmt.Errorf("cannot parse %s: %w", name, err)
		return source
	}
	for _, key := range trackedKeys {
		if _, ok := values[key]; ok {
			source.Keys = append(source.Keys, key)
		}
	}
	_ = json.Unmarshal(values["model"], &source.Model)
	return source
}

// stripJSONC turns JSONC (JSON with // and /* */ comments and trailing commas, which OpenCode accepts)
// into plain JSON, leaving string contents such as "http://..." alone.
func stripJSONC(body []byte) []byte {
	out := make([]byte, 0, len(body))
	inString := false
	for i := 0; i < len(body); i++ {
		c := body[i]
		switch {
		case inString:
			out = append(out, c)
			if c == '\\' && i+1 < len(body) {
				i++
				out = append(out, body[i])
			} else if c == '"' {
				inString = false
			}
		case c == '"':
			inString = true
			out = append(out, c)
		case c == '/' && i+1 < len(body) && body[i+1] == '/':
			for i < len(body) && body[i] != '\n' {
				i++
			}
			out = append(out, '\n')
		case c == '/' && i+1 < len(body) && body[i+1] == '*':
			for i += 2; i+1 < len(body) && !(body[i] == '*' && body[i+1] == '/'); i++ {
			}
			i++
		case c == '}' || c == ']':
			// Drop a trailing comma before the closing bracket.
			j := len(out) - 1
			for j >= 0 && (out[j] == ' ' || out[j] == '\t' || out[j] == '\n' || out[j] == '\r') {
				j--
			}
			if j >= 0 && out[j] == ',' {
				out = append(out[:j], out[j+1:]...)
			}
			out = append(out, c)
		default:
			out = append(out, c)
		}
	}
	return out
}

func displayHome(path string) string {
	if home := config.HomeDir(); home != "" && strings.HasPrefix(path, home+string(filepath.Separator)) {
		return "~" + filepath.ToSlash(strings.TrimPrefix(path, home))
	}
	return path
}
