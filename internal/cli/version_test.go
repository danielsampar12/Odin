package cli

import "testing"

func TestDisplayVersion(t *testing.T) {
	testCases := []struct {
		name    string
		release string
		module  string
		want    string
	}{
		{name: "release tag beats the module version", release: "v0.3.0", module: "v0.2.0", want: "v0.3.0"},
		{name: "go install of a tag", module: "v0.3.0", want: "v0.3.0"},
		{name: "build without version control", module: "(devel)", want: "dev"},
		{name: "no build info", want: "dev"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := displayVersion(tc.release, tc.module); got != tc.want {
				t.Fatalf("displayVersion(%q, %q) = %q, want %q", tc.release, tc.module, got, tc.want)
			}
		})
	}
}
