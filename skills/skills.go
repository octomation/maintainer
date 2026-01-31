// Package skills holds the agent skills that ship with maintainer.
package skills

import (
	"embed"
	"io/fs"
)

//go:generate go run ../internal/skill/gen maintainer/references/cli.md

//go:embed maintainer
var files embed.FS

// Maintainer returns the maintainer skill with SKILL.md at its root.
func Maintainer() fs.FS {
	sub, err := fs.Sub(files, "maintainer")
	if err != nil {
		panic(err)
	}
	return sub
}
