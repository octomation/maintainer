package skill

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// Frontmatter is the parsed header of SKILL.md. Values are strings, or
// map[string]string for one level of nesting such as metadata.
type Frontmatter map[string]any

// String returns a top-level string value.
func (f Frontmatter) String(key string) (string, bool) {
	s, ok := f[key].(string)
	return s, ok
}

// Map returns a nested map such as metadata.
func (f Frontmatter) Map(key string) (map[string]string, bool) {
	m, ok := f[key].(map[string]string)
	return m, ok
}

var (
	entry       = regexp.MustCompile(`^([A-Za-z0-9_.-]+):(?:[ ]+(.*))?$`)
	blockStyle  = regexp.MustCompile(`^[>|]-?$`)
	doubleQuote = regexp.MustCompile(`^"(?:[^"\\]|\\.)*"$`)
	singleQuote = regexp.MustCompile(`^'(?:[^']|'')*'$`)
	indicator   = regexp.MustCompile("^[\\[\\]{}&*!|>@`%,?#-]")
	dashWord    = regexp.MustCompile(`^-\S`)
	inlineMark  = regexp.MustCompile(`:\s|\s#`)
	nonString   = regexp.MustCompile(`(?i)^(?:~|null|true|false|yes|no|on|off|y|n)$`)
	number      = regexp.MustCompile(`(?i)^[-+]?(?:\d[\d_]*(?:\.\d*)?(?:e[-+]?\d+)?|\.\d+|0x[\da-f]+|0o[0-7]+|\.inf|\.nan)$`)
)

// ParseFrontmatter reads the YAML subset SKILL.md may use and rejects the rest
// instead of guessing, exactly as the octolab/skills catalog does:
//
//	key: plain scalar          a string that YAML would also read as a string
//	key: "double" | 'single'   quoted strings
//	key: >- | > | |- | |       block scalars indented by two spaces
//	key:                       a map of the same key-value lines, indented by
//	  sub: value               two spaces, one level deep
//
// Duplicate keys, flow collections, anchors, tags, and plain scalars YAML
// would type as booleans, nulls or numbers are errors. It returns the
// frontmatter and the body after it.
func ParseFrontmatter(text string) (Frontmatter, string, error) {
	lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	if lines[0] != "---" {
		return nil, "", fmt.Errorf("SKILL.md must start with a --- frontmatter line")
	}
	end := -1
	for i := 1; i < len(lines); i++ {
		if lines[i] == "---" {
			end = i
			break
		}
	}
	if end < 0 {
		return nil, "", fmt.Errorf("SKILL.md frontmatter has no closing --- line")
	}
	p := parser{lines: lines[1:end]}
	data := Frontmatter{}
	if err := p.entries(0, data, ""); err != nil {
		return nil, "", err
	}
	return data, strings.Join(lines[end+1:], "\n"), nil
}

type parser struct {
	lines []string
	i     int
}

func indent(line string) int { return len(line) - len(strings.TrimLeft(line, " \t")) }

func blank(line string) bool { return strings.TrimSpace(line) == "" }

func (p *parser) entries(base int, target map[string]any, where string) error {
	for p.i < len(p.lines) {
		line := p.lines[p.i]
		if blank(line) || strings.HasPrefix(strings.TrimLeft(line, " \t"), "#") {
			p.i++
			continue
		}
		if indent(line) < base {
			return nil
		}
		if indent(line) != base {
			return fmt.Errorf("%s: inconsistent indentation: %s", where, strings.TrimSpace(line))
		}
		m := entry.FindStringSubmatch(line[base:])
		if m == nil {
			return fmt.Errorf("%s: unsupported line: %s", where, strings.TrimSpace(line))
		}
		key, rest := m[1], m[2]
		at := key
		if where != "" {
			at = where + "." + key
		}
		if _, dup := target[key]; dup {
			return fmt.Errorf("%s: duplicate key", at)
		}
		p.i++
		switch {
		case blockStyle.MatchString(rest):
			v, err := p.block(rest, base, at)
			if err != nil {
				return err
			}
			target[key] = v
		case rest == "":
			if base > 0 {
				return fmt.Errorf("%s: only one level of nesting is supported", at)
			}
			nested := map[string]any{}
			if err := p.entries(base+2, nested, at); err != nil {
				return err
			}
			if len(nested) == 0 {
				return fmt.Errorf("%s: empty value", at)
			}
			flat := make(map[string]string, len(nested))
			for k, v := range nested {
				flat[k] = v.(string)
			}
			target[key] = flat
		default:
			v, err := scalar(rest, at)
			if err != nil {
				return err
			}
			target[key] = v
		}
	}
	return nil
}

func (p *parser) block(style string, base int, where string) (string, error) {
	var out []string
	for p.i < len(p.lines) && (blank(p.lines[p.i]) || indent(p.lines[p.i]) > base) {
		line := p.lines[p.i]
		switch {
		case blank(line):
			out = append(out, "")
		case indent(line) != base+2:
			return "", fmt.Errorf("%s: block lines must be indented by %d spaces", where, base+2)
		default:
			out = append(out, line[base+2:])
		}
		p.i++
	}
	for len(out) > 0 && out[len(out)-1] == "" {
		out = out[:len(out)-1]
	}
	if len(out) == 0 {
		return "", fmt.Errorf("%s: empty block scalar", where)
	}
	joined := strings.Join(out, "\n")
	if strings.HasPrefix(style, ">") {
		// Folding joins adjacent lines with a space; each empty line between
		// them stands for one newline.
		var b strings.Builder
		empty := 0
		for _, line := range out {
			if line == "" {
				empty++
				continue
			}
			switch {
			case b.Len() > 0 && empty > 0, b.Len() == 0:
				b.WriteString(strings.Repeat("\n", empty))
			default:
				b.WriteString(" ")
			}
			b.WriteString(line)
			empty = 0
		}
		joined = b.String()
	}
	if strings.HasSuffix(style, "-") {
		return joined, nil
	}
	return joined + "\n", nil
}

func scalar(raw, where string) (string, error) {
	v := strings.TrimSpace(raw)
	switch {
	case strings.HasPrefix(v, `"`):
		if !doubleQuote.MatchString(v) {
			return "", fmt.Errorf("%s: unterminated or trailing double-quoted string", where)
		}
		var s string
		if err := json.Unmarshal([]byte(v), &s); err != nil {
			return "", fmt.Errorf("%s: %w", where, err)
		}
		return s, nil
	case strings.HasPrefix(v, "'"):
		if !singleQuote.MatchString(v) {
			return "", fmt.Errorf("%s: unterminated or trailing single-quoted string", where)
		}
		return strings.ReplaceAll(v[1:len(v)-1], "''", "'"), nil
	case indicator.MatchString(v) && !dashWord.MatchString(v):
		return "", fmt.Errorf("%s: quote values that start with %c", where, v[0])
	case inlineMark.MatchString(v):
		return "", fmt.Errorf(`%s: quote values that contain ": " or " #"`, where)
	case nonString.MatchString(v):
		return "", fmt.Errorf(`%s: quote %q, YAML reads it as a non-string`, where, v)
	case number.MatchString(v):
		return "", fmt.Errorf(`%s: quote %q, YAML reads it as a number`, where, v)
	}
	return v, nil
}
