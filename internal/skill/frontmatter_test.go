package skill_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "go.octolab.org/toolset/maintainer/internal/skill"
)

const skillMD = `---
name: maintainer
description: >-
  Fetch checkouts with the maintainer CLI.
  Use when the user wants to clone repositories.
license: MIT
compatibility: Requires the maintainer binary.
metadata:
  author: octomation
  version: "0.2.0"
  tool: maintainer
  tool-version-range: ">=0.2.0 <0.3.0"
---

# maintainer

See [workflows](references/workflows.md).
`

func TestParseFrontmatter(t *testing.T) {
	fm, body, err := ParseFrontmatter(skillMD)
	require.NoError(t, err)
	name, _ := fm.String("name")
	assert.Equal(t, "maintainer", name)
	description, _ := fm.String("description")
	assert.Equal(t, "Fetch checkouts with the maintainer CLI. Use when the user wants to clone repositories.", description)
	meta, ok := fm.Map("metadata")
	require.True(t, ok)
	assert.Equal(t, map[string]string{
		"author":             "octomation",
		"version":            "0.2.0",
		"tool":               "maintainer",
		"tool-version-range": ">=0.2.0 <0.3.0",
	}, meta)
	assert.Contains(t, body, "# maintainer")
}

func TestParseFrontmatter_blocks(t *testing.T) {
	field := func(yaml string) string {
		fm, _, err := ParseFrontmatter("---\n" + yaml + "\n---\n")
		require.NoError(t, err, yaml)
		s, _ := fm.String("d")
		return s
	}
	assert.Equal(t, "a\nb", field("d: >-\n  a\n\n  b"))
	assert.Equal(t, "a\n\nb", field("d: >-\n  a\n\n\n  b"))
	assert.Equal(t, "a b\n", field("d: >\n  a\n  b"))
	assert.Equal(t, "a\n\nb", field("d: |-\n  a\n\n  b"))
	assert.Equal(t, "it's", field("d: 'it''s'"))
	assert.Equal(t, "a\tb", field(`d: "a\tb"`))
}

func TestParseFrontmatter_rejects(t *testing.T) {
	tests := map[string]string{
		"no frontmatter":                    "must start",
		"---\nname: x\n":                    "closing",
		"description: [invalid, sequence]":  "quote values that start with [",
		"tags: {a: b}":                      "quote values that start with {",
		"metadata:\n  author: true":         "non-string",
		"metadata:\n  version: 1.0":         "number",
		"license: 'MIT":                     "unterminated",
		`license: "MIT" extra`:              "unterminated or trailing",
		"name: a\nname: b":                  "duplicate key",
		"metadata:\n  a: x\n   b: y":        "inconsistent indentation",
		"metadata:\n  nested:\n    deep: x": "one level",
		"range: >=1.0.0 <2.0.0":             "start with >",
		"note: a: b":                        `contain ": "`,
		"list:\n  - a":                      "unsupported line",
		"d: >-\n   a":                       "indented by 2 spaces",
		"metadata:":                         "empty value",
		"d: |-":                             "empty block scalar",
	}
	for yaml, problem := range tests {
		text := yaml
		if yaml != "no frontmatter" && yaml != "---\nname: x\n" {
			text = "---\n" + yaml + "\n---\n"
		}
		_, _, err := ParseFrontmatter(text)
		if assert.Error(t, err, yaml) {
			assert.Contains(t, err.Error(), problem, yaml)
		}
	}
}
