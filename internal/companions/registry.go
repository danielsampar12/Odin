package companions

import (
	"strings"

	"github.com/danielsampar12/odin/roles"
)

// Companion is an OpenCode agent Odin generates; its prompt is roles/<Key>.md.
type Companion struct {
	Key         string
	Name        string
	Description string
	ReadOnly    bool // cannot edit files, run arbitrary commands or hand work to subagents
}

// Registry lists the companions in the order `odin companion list` shows them; the first is the default.
var Registry = []Companion{
	{Key: "baldur", Name: "Baldur", Description: "Pair programmer: makes the changes you ask for, runs the tests and explains tradeoffs"},
	{Key: "tyr", Name: "Tyr", Description: "Reviewer: reads code and diffs, runs tests and reports problems; cannot edit files", ReadOnly: true},
}

func Default() Companion {
	return Registry[0]
}

// Lookup finds a companion by key. An unknown key, such as a v1 companion that Odin v2 no longer ships,
// returns the default companion and false.
func Lookup(key string) (Companion, bool) {
	for _, companion := range Registry {
		if strings.EqualFold(companion.Key, key) {
			return companion, true
		}
	}
	return Default(), false
}

// Prompt is the companion's system prompt for OpenCode. It replaces OpenCode's built-in prompt for that agent.
func (c Companion) Prompt() string {
	body, err := roles.FS.ReadFile(c.Key + ".md")
	if err != nil {
		panic("roles/" + c.Key + ".md is not embedded: " + err.Error()) // every Registry entry ships a role file
	}
	return string(body)
}

func DisplayName(key string) string {
	companion, ok := Lookup(key)
	if !ok {
		return key
	}
	return companion.Name
}
