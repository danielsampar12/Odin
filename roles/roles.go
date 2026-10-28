// Package roles embeds the companion prompts so the odin binary carries them.
package roles

import "embed"

//go:embed *.md
var FS embed.FS
