// Command gen writes the skill's command reference from the command tree:
//
//	go run ./internal/skill/gen <file>
package main

import (
	"fmt"
	"os"

	"go.octolab.org/toolset/maintainer/internal/command"
	"go.octolab.org/toolset/maintainer/internal/skill"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: gen <file>")
		os.Exit(2)
	}
	if err := os.WriteFile(os.Args[1], []byte(skill.Reference(command.New())), 0o644); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
