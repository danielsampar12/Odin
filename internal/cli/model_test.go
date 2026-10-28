package cli

import (
	"strings"
	"testing"
)

func TestModelRecommendRejectsUnknownUseCase(t *testing.T) {
	cmd := NewRootCmd()
	cmd.SetArgs([]string{"model", "recommend", "--use-case", "foo"})

	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), `unknown --use-case "foo"`) {
		t.Fatalf("err = %v, want unknown --use-case error", err)
	}
}
