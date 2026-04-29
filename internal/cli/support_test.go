package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestConfirmAction(t *testing.T) {
	testCases := []struct {
		name   string
		input  string
		want   bool
		prompt string
	}{
		{
			name:   "yes short",
			input:  "y\n",
			want:   true,
			prompt: "Continue? [y/N]: ",
		},
		{
			name:   "yes long",
			input:  "yes\n",
			want:   true,
			prompt: "Continue? [y/N]: ",
		},
		{
			name:   "default no",
			input:  "\n",
			want:   false,
			prompt: "Continue? [y/N]: ",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var output bytes.Buffer
			got, err := confirmAction(strings.NewReader(tc.input), &output, tc.prompt)
			if err != nil {
				t.Fatalf("confirmAction error = %v", err)
			}
			if got != tc.want {
				t.Fatalf("confirmAction = %v, want %v", got, tc.want)
			}
			if output.String() != tc.prompt {
				t.Fatalf("prompt output = %q, want %q", output.String(), tc.prompt)
			}
		})
	}
}
