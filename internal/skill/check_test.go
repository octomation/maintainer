package skill_test

import (
	"strings"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"

	. "go.octolab.org/toolset/maintainer/internal/skill"
)

func skillFS(text string) fstest.MapFS {
	return fstest.MapFS{
		"SKILL.md":                {Data: []byte(text)},
		"references/workflows.md": {Data: []byte("Back to [the skill](../SKILL.md#maintainer).\n")},
	}
}

func TestCheck(t *testing.T) {
	assert.Empty(t, Check("maintainer", skillFS(skillMD)))

	bad := strings.NewReplacer(
		"name: maintainer", "name: other",
		`version: "0.2.0"`, `version: "0.3.0"`,
		"references/workflows.md", "references/missing.md",
	).Replace(skillMD)
	var problems []string
	for _, err := range Check("maintainer", skillFS(bad)) {
		problems = append(problems, err.Error())
	}
	joined := strings.Join(problems, "\n")
	assert.Contains(t, joined, "name: must be maintainer")
	assert.Contains(t, joined, "tool-version-range: must start at the release")
	assert.Contains(t, joined, "broken link: references/missing.md")

	outside := skillFS(strings.Replace(skillMD, "references/workflows.md", "../README.md", 1))
	assert.ErrorContains(t, Check("maintainer", outside)[0], "link leaves the skill")
	assert.ErrorContains(t, Check("maintainer", fstest.MapFS{"README.md": {}})[0], "SKILL.md: missing")
}
