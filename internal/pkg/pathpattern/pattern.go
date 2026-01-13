// Package pathpattern compiles reversible checkout templates into bounded
// directory matchers. Unsupported template constructs fail closed.
package pathpattern

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
	"text/template/parse"
)

type Pattern struct{ parts []*regexp.Regexp }

// Compile accepts literal path text and scalar fields, optionally lower/upper.
// owner binds owner overrides without consulting a remote or state.
func Compile(source, owner string) (*Pattern, error) {
	t, err := template.New("scope").Funcs(template.FuncMap{"lower": strings.ToLower, "upper": strings.ToUpper}).Parse(source)
	if err != nil {
		return nil, err
	}
	var expression strings.Builder
	for _, node := range t.Tree.Root.Nodes {
		switch n := node.(type) {
		case *parse.TextNode:
			expression.WriteString(regexp.QuoteMeta(string(n.Text)))
		case *parse.ActionNode:
			field, transform, err := scalar(n.Pipe)
			if err != nil {
				return nil, fmt.Errorf("workspace path %q: %w", source, err)
			}
			var values []string
			switch field {
			case "Visibility":
				values = []string{"public", "private", "internal"}
			case "IsFork", "IsTemplate", "IsArchived":
				values = []string{"true", "false"}
			case "Owner":
				if owner != "" {
					values = []string{owner}
				}
			case "Repo", "DefaultBranch":
			default:
				return nil, fmt.Errorf("workspace path: unsupported field %q", field)
			}
			if len(values) == 0 {
				// The sentinel is expanded after splitting literal separators.
				expression.WriteString("\x00")
			} else {
				for i, value := range values {
					if transform == "lower" {
						value = strings.ToLower(value)
					} else if transform == "upper" {
						value = strings.ToUpper(value)
					}
					values[i] = regexp.QuoteMeta(value)
				}
				expression.WriteString("(?:" + strings.Join(values, "|") + ")")
			}
		default:
			return nil, fmt.Errorf("workspace path %q: only literal text and scalar fields are supported", source)
		}
	}
	if source == "" || strings.ContainsAny(source, "\x00\\") || filepath.IsAbs(source) || strings.HasPrefix(source, "~") {
		return nil, fmt.Errorf("workspace path must be a nonempty relative template")
	}
	p := &Pattern{}
	for _, part := range strings.Split(expression.String(), "/") {
		if part == "" || part == `\.` || part == `\.\.` {
			return nil, fmt.Errorf("workspace path contains an empty, . or .. component")
		}
		r, err := regexp.Compile("^(?:" + strings.ReplaceAll(part, "\x00", "[^/]+") + ")$")
		if err != nil {
			return nil, err
		}
		p.parts = append(p.parts, r)
	}
	return p, nil
}

func scalar(pipe *parse.PipeNode) (string, string, error) {
	if len(pipe.Decl) != 0 || len(pipe.Cmds) < 1 || len(pipe.Cmds) > 2 {
		return "", "", fmt.Errorf("unsupported template pipeline")
	}
	args := pipe.Cmds[0].Args
	transform := ""
	if len(args) == 2 {
		id, ok := args[0].(*parse.IdentifierNode)
		if !ok || (id.Ident != "lower" && id.Ident != "upper") {
			return "", "", fmt.Errorf("unsupported template function")
		}
		transform, args = id.Ident, args[1:]
	}
	if len(args) != 1 {
		return "", "", fmt.Errorf("expected one scalar field")
	}
	f, ok := args[0].(*parse.FieldNode)
	if !ok || len(f.Ident) != 1 {
		return "", "", fmt.Errorf("expected one scalar field")
	}
	if len(pipe.Cmds) == 2 {
		last := pipe.Cmds[1].Args
		if transform != "" || len(last) != 1 {
			return "", "", fmt.Errorf("unsupported template pipeline")
		}
		id, ok := last[0].(*parse.IdentifierNode)
		if !ok || (id.Ident != "lower" && id.Ident != "upper") {
			return "", "", fmt.Errorf("unsupported template function")
		}
		transform = id.Ident
	}
	return f.Ident[0], transform, nil
}

// Match tests complete checkout paths; Prefix tests whether walking can still
// reach a match. Paths are relative to the workspace root.
func (p *Pattern) Match(path string) bool  { return p.match(path, false) }
func (p *Pattern) Prefix(path string) bool { return p.match(path, true) }

func (p *Pattern) match(path string, prefix bool) bool {
	if path == "." || path == "" {
		return prefix
	}
	parts := strings.Split(filepath.ToSlash(path), "/")
	if len(parts) > len(p.parts) || (!prefix && len(parts) != len(p.parts)) {
		return false
	}
	for i, part := range parts {
		if part == "." || part == ".." || !p.parts[i].MatchString(part) {
			return false
		}
	}
	return true
}
