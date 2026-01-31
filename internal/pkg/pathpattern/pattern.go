// Package pathpattern compiles reversible checkout templates into bounded
// directory matchers. Unsupported template constructs fail closed.
package pathpattern

import (
	"fmt"
	"maps"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"text/template"
	"text/template/parse"
)

type Pattern struct {
	parts      []*regexp.Regexp
	components [][]token // the template split into path components
}

// token is literal text or a field. A finite field lists its values; a free
// field (.Repo, .DefaultBranch, an unbound .Owner) takes any component text.
type token struct {
	text      string
	field     string
	transform string
	values    []string
}

// Compile accepts literal path text and scalar fields, optionally lower/upper.
// owner binds owner overrides without consulting a remote or state; a bound
// .Owner matches any letter case, as GitHub logins compare.
func Compile(source, owner string) (*Pattern, error) {
	t, err := template.New("scope").Funcs(template.FuncMap{"lower": strings.ToLower, "upper": strings.ToUpper}).Parse(source)
	if err != nil {
		return nil, err
	}
	var expression strings.Builder
	var tokens []token
	for _, node := range t.Root.Nodes {
		switch n := node.(type) {
		case *parse.TextNode:
			expression.WriteString(regexp.QuoteMeta(string(n.Text)))
			tokens = append(tokens, token{text: string(n.Text)})
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
			tokens = append(tokens, token{field: field, transform: transform, values: slices.Clone(values)})
			if len(values) == 0 {
				// The sentinel is expanded after splitting literal separators.
				expression.WriteString("\x00")
			} else {
				for i, value := range values {
					switch transform {
					case "lower":
						value = strings.ToLower(value)
					case "upper":
						value = strings.ToUpper(value)
					}
					values[i] = regexp.QuoteMeta(value)
				}
				flags := ""
				if field == "Owner" && transform == "" {
					flags = "(?i)" // rendered with GitHub's spelling of the login
				}
				expression.WriteString("(?:" + flags + strings.Join(values, "|") + ")")
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
	p.components = split(tokens)
	return p, nil
}

// split groups tokens into path components at the separators of literal text.
func split(tokens []token) [][]token {
	components := [][]token{nil}
	for _, tok := range tokens {
		if tok.field != "" {
			components[len(components)-1] = append(components[len(components)-1], tok)
			continue
		}
		for i, piece := range strings.Split(tok.text, "/") {
			if i > 0 {
				components = append(components, nil)
			}
			if piece != "" {
				components[len(components)-1] = append(components[len(components)-1], token{text: piece})
			}
		}
	}
	return components
}

// Nested reports a demonstrable ancestor relation between the paths of p and
// q: it searches for one consistent assignment of the template fields of both
// under which the shorter path is a prefix of the longer one. A repeated field
// keeps one value, and an unbound .Owner never takes a reserved owner, which
// has its own template. Patterns of equal depth never nest. When no witness is
// found, the planner still checks the actual targets.
func Nested(p, q *Pattern, reserved []string) bool {
	if len(p.components) == len(q.components) {
		return false
	}
	depth := min(len(p.components), len(q.components))
	for _, a := range p.assignments() {
		for _, b := range q.assignments() {
			u := unifier{sides: [2]map[string]bound{{}, {}}, pairs: make([][2][]part, depth), reserved: reserved}
			for i := range depth {
				u.pairs[i] = [2][]part{render(p.components[i], a), render(q.components[i], b)}
			}
			if u.solve() {
				return true
			}
		}
	}
	return false
}

// assignments enumerates every combination of the finite fields' values.
func (p *Pattern) assignments() []map[string]string {
	result := []map[string]string{{}}
	for _, component := range p.components {
		for _, tok := range component {
			if len(tok.values) == 0 {
				continue
			}
			if _, done := result[0][tok.field]; done {
				continue
			}
			next := make([]map[string]string, 0, len(result)*len(tok.values))
			for _, assignment := range result {
				for _, value := range tok.values {
					extended := maps.Clone(assignment)
					extended[tok.field] = value
					next = append(next, extended)
				}
			}
			result = next
		}
	}
	return result
}

// part is rendered literal text or a free field.
type part struct{ text, field, transform string }

func transform(name, value string) string {
	switch name {
	case "lower":
		return strings.ToLower(value)
	case "upper":
		return strings.ToUpper(value)
	}
	return value
}

func render(component []token, assignment map[string]string) []part {
	var parts []part
	for _, tok := range component {
		if tok.field != "" && len(tok.values) == 0 {
			parts = append(parts, part{field: tok.field, transform: tok.transform})
			continue
		}
		text := tok.text
		if tok.field != "" {
			text = transform(tok.transform, assignment[tok.field])
		}
		if n := len(parts); n > 0 && parts[n-1].field == "" {
			parts[n-1].text += text
		} else {
			parts = append(parts, part{text: text})
		}
	}
	return parts
}

// bound is the value chosen for a free field, as rendered with its transform.
type bound struct{ transform, value string }

type unifier struct {
	sides    [2]map[string]bound // free fields of each pattern
	pairs    [][2][]part
	reserved []string
	fresh    int
}

func (u *unifier) set(side int, field, how, value string) bool {
	if value == "" || strings.Contains(value, "/") || transform(how, value) != value {
		return false
	}
	if field == "Owner" && slices.ContainsFunc(u.reserved, func(owner string) bool { return strings.EqualFold(owner, value) }) {
		return false
	}
	if b, ok := u.sides[side][field]; ok {
		return b == bound{how, value}
	}
	u.sides[side][field] = bound{how, value}
	return true
}

// text returns the component text when every free field is already bound.
func (u *unifier) text(side int, parts []part) (string, bool) {
	var b strings.Builder
	for _, p := range parts {
		if p.field == "" {
			b.WriteString(p.text)
			continue
		}
		v, ok := u.sides[side][p.field]
		if !ok || v.transform != p.transform {
			return "", false
		}
		b.WriteString(v.value)
	}
	return b.String(), true
}

// capture binds the free fields of parts so that they render text.
func (u *unifier) capture(side int, parts []part, text string) bool {
	var expression strings.Builder
	var fields []part
	for _, p := range parts {
		if p.field == "" {
			expression.WriteString(regexp.QuoteMeta(p.text))
			continue
		}
		if v, ok := u.sides[side][p.field]; ok && v.transform == p.transform {
			expression.WriteString(regexp.QuoteMeta(v.value))
			continue
		}
		expression.WriteString("([^/]+)")
		fields = append(fields, p)
	}
	match := regexp.MustCompile("^" + expression.String() + "$").FindStringSubmatch(text)
	if match == nil {
		return false
	}
	for i, p := range fields {
		if !u.set(side, p.field, p.transform, match[i+1]) {
			return false
		}
	}
	return true
}

// solve matches components where one side is already text first, then
// gives a free field alone in its component the other side's text, with
// fresh values for that side's own free fields; anything else is no witness.
func (u *unifier) solve() bool {
	pending := u.pairs
	for len(pending) > 0 {
		var rest [][2][]part
		for _, pair := range pending {
			a, aok := u.text(0, pair[0])
			b, bok := u.text(1, pair[1])
			switch {
			case aok && bok:
				if a != b {
					return false
				}
			case aok:
				if !u.capture(1, pair[1], a) {
					return false
				}
			case bok:
				if !u.capture(0, pair[0], b) {
					return false
				}
			default:
				rest = append(rest, pair)
			}
		}
		if len(rest) == len(pending) && !slices.ContainsFunc(rest, u.guess) {
			return false
		}
		pending = rest
	}
	return true
}

func (u *unifier) guess(pair [2][]part) bool {
	for side := range 2 {
		alone, other := pair[side], pair[1-side]
		if len(alone) != 1 || alone[0].field == "" {
			continue
		}
		for _, p := range other {
			if _, ok := u.sides[1-side][p.field]; p.field != "" && !ok {
				u.fresh++
				if !u.set(1-side, p.field, p.transform, fmt.Sprintf("_%d", u.fresh)) {
					return false
				}
			}
		}
		text, ok := u.text(1-side, other)
		return ok && u.set(side, alone[0].field, alone[0].transform, text)
	}
	return false
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
