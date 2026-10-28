package companions

import "testing"

func TestEveryCompanionHasAPrompt(t *testing.T) {
	for _, companion := range Registry {
		if len(companion.Prompt()) < 100 {
			t.Fatalf("%s prompt is missing or too short", companion.Key)
		}
	}
	if companion, ok := Lookup("freya"); ok || companion.Key != "baldur" {
		t.Fatalf("Lookup(freya) = %v, %v; want the default and false", companion.Key, ok)
	}
}
