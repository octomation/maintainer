// Package skills holds the agent skills that ship with maintainer.
package skills

import (
	"embed"
	"io/fs"
)

//go:generate go run ../internal/skill/gen maintainer/references/cli.md

//go:embed maintainer review
var files embed.FS

// Names lists the skills that ship with maintainer, the main skill first.
// Each is a directory here and in the catalog's plugin.
var Names = []string{"maintainer", "review"}

// Maintainer returns the maintainer skill with SKILL.md at its root.
func Maintainer() fs.FS { return Skill("maintainer") }

// Skill returns the named skill with SKILL.md at its root.
func Skill(name string) fs.FS {
	sub, err := fs.Sub(files, name)
	if err != nil {
		panic(err)
	}
	return sub
}

// All returns the skills in the order of Names.
func All() []fs.FS {
	list := make([]fs.FS, 0, len(Names))
	for _, name := range Names {
		list = append(list, Skill(name))
	}
	return list
}
