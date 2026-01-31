// Package skill implements the contract of the agent skill that ships with
// maintainer: the SKILL.md frontmatter, the compatibility policy, the payload
// digest and the generated command reference. The rules match the
// octolab/skills catalog, which validates the same skill on publication.
package skill

import (
	"fmt"
	"io/fs"
	"path"
	"regexp"
	"strings"
)

var (
	link   = regexp.MustCompile(`\]\(([^)\s]+)\)`)
	scheme = regexp.MustCompile(`(?i)^[a-z][a-z0-9+.-]*:`)
)

// Check validates the skill against the Agent Skills spec and the OctoLab
// metadata contract and returns every problem it finds.
func Check(tool string, fsys fs.FS) []error {
	files, err := Files(fsys)
	if err != nil {
		return []error{err}
	}
	text, err := fs.ReadFile(fsys, "SKILL.md")
	if err != nil {
		return []error{fmt.Errorf("SKILL.md: missing")}
	}
	fm, _, err := ParseFrontmatter(string(text))
	if err != nil {
		return []error{fmt.Errorf("SKILL.md: %w", err)}
	}

	var errs []error
	report := func(format string, args ...any) { errs = append(errs, fmt.Errorf(format, args...)) }

	if name, _ := fm.String("name"); name != tool {
		report("SKILL.md name: must be %s", tool)
	}
	if d, ok := fm.String("description"); !ok || strings.TrimSpace(d) == "" || len([]rune(d)) > 1024 {
		report("SKILL.md description: required, at most 1024 characters")
	}
	if _, present := fm["compatibility"]; present {
		if c, ok := fm.String("compatibility"); !ok || len([]rune(c)) > 500 {
			report("SKILL.md compatibility: at most 500 characters")
		}
	}
	meta, ok := fm.Map("metadata")
	if !ok {
		report("SKILL.md metadata: required")
		return errs
	}
	if meta["tool"] != tool {
		report("SKILL.md metadata.tool: must be %s", tool)
	}
	if v, ok := ParseVersion(meta["version"]); !ok || strings.HasPrefix(meta["version"], "v") {
		report("SKILL.md metadata.version: must be a semantic version")
	} else if err := CheckRange(v, meta["tool-version-range"]); err != nil {
		report("SKILL.md metadata.tool-version-range: %v", err)
	}

	exists := make(map[string]bool)
	for _, f := range files {
		for dir := f; dir != "."; dir = path.Dir(dir) {
			exists[dir] = true
		}
	}
	for _, f := range files {
		if !strings.HasSuffix(f, ".md") {
			continue
		}
		data, err := fs.ReadFile(fsys, f)
		if err != nil {
			return append(errs, err)
		}
		for _, m := range link.FindAllStringSubmatch(string(data), -1) {
			target := m[1]
			if scheme.MatchString(target) || strings.HasPrefix(target, "#") {
				continue
			}
			rel, _, _ := strings.Cut(target, "#")
			resolved := path.Join(path.Dir(f), rel)
			switch {
			case resolved == ".." || strings.HasPrefix(resolved, "../") || path.IsAbs(rel):
				report("%s: link leaves the skill: %s", f, target)
			case !exists[resolved]:
				report("%s: broken link: %s", f, target)
			}
		}
	}
	return errs
}
