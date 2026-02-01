// Package mirror reads the issue mirrors of a checkout, `.github/issues` and
// `.github/roadmap`, in their three frontmatter generations, and patches the
// triage labels of an issue file byte for byte.
package mirror

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

// Document is a mirror file split into its frontmatter and body. Raw keeps
// every byte: the writer patches it in place.
type Document struct {
	Path    string
	Raw     []byte
	Version string // "sha256-<hex>" of Raw
	Fields  map[string]any
	Body    string
	CRLF    bool
	// frontmatter is the byte range of the lines between the two delimiters.
	frontStart, frontEnd int
}

// ErrNoFrontmatter reports a file that does not start with a frontmatter.
var ErrNoFrontmatter = errors.New("no frontmatter")

// Read loads a mirror file.
func Read(path string) (*Document, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	doc, err := Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", filepath.Base(path), err)
	}
	doc.Path = path
	return doc, nil
}

// Parse splits raw into frontmatter and body and decodes the frontmatter.
func Parse(raw []byte) (*Document, error) {
	doc := &Document{Raw: raw, Version: Hash(raw)}
	nl := "\n"
	if i := bytes.IndexByte(raw, '\n'); i > 0 && raw[i-1] == '\r' {
		nl, doc.CRLF = "\r\n", true
	}
	open := "---" + nl
	if !bytes.HasPrefix(raw, []byte(open)) {
		return nil, ErrNoFrontmatter
	}
	doc.frontStart = len(open)
	rest := raw[doc.frontStart:]
	// The closing delimiter is a line of its own; it may end the file.
	end := -1
	if bytes.HasPrefix(rest, []byte("---"+nl)) || string(rest) == "---" {
		end = 0
	} else if i := bytes.Index(rest, []byte(nl+"---"+nl)); i >= 0 {
		end = i + len(nl)
	} else if bytes.HasSuffix(rest, []byte(nl+"---")) {
		end = len(rest) - 3
	}
	if end < 0 {
		return nil, fmt.Errorf("%w: the closing --- is missing", ErrNoFrontmatter)
	}
	doc.frontEnd = doc.frontStart + end
	after := doc.frontEnd + 3
	if after < len(raw) {
		after += len(nl)
	}
	if after > len(raw) {
		after = len(raw)
	}
	doc.Body = string(raw[after:])
	var fields map[string]any
	if err := yaml.Unmarshal(raw[doc.frontStart:doc.frontEnd], &fields); err != nil {
		return nil, fmt.Errorf("frontmatter: %w", err)
	}
	if fields == nil {
		fields = map[string]any{}
	}
	doc.Fields = normalizeMap(fields)
	return doc, nil
}

// Frontmatter returns the bytes between the delimiters.
func (d *Document) Frontmatter() []byte { return d.Raw[d.frontStart:d.frontEnd] }

// Has reports whether the frontmatter has the key, even with a null value.
func (d *Document) Has(key string) bool {
	_, ok := d.Fields[key]
	return ok
}

// String returns the key as a string: a number is formatted, null is "".
func (d *Document) String(key string) string {
	switch v := d.Fields[key].(type) {
	case nil:
		return ""
	case string:
		return v
	case int:
		return strconv.Itoa(v)
	case int64:
		return strconv.FormatInt(v, 10)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(v)
	default:
		return ""
	}
}

// Int returns the key as a number when it is one.
func (d *Document) Int(key string) (int, bool) {
	switch v := d.Fields[key].(type) {
	case int:
		return v, true
	case int64:
		return int(v), true
	case float64:
		if v == float64(int(v)) {
			return int(v), true
		}
	case string:
		if n, err := strconv.Atoi(strings.TrimSpace(v)); err == nil {
			return n, true
		}
	}
	return 0, false
}

// Strings returns the key as a list of strings; scalars become one item.
func (d *Document) Strings(key string) []string {
	switch v := d.Fields[key].(type) {
	case nil:
		return nil
	case []any:
		out := make([]string, 0, len(v))
		for _, item := range v {
			switch s := item.(type) {
			case string:
				out = append(out, s)
			case nil:
			default:
				out = append(out, fmt.Sprint(s))
			}
		}
		return out
	case string:
		if v == "" {
			return nil
		}
		return []string{v}
	default:
		return nil
	}
}

// Map returns the key as a map when it is one.
func (d *Document) Map(key string) map[string]any {
	if m, ok := d.Fields[key].(map[string]any); ok {
		return m
	}
	return nil
}

// Time parses the key as an RFC 3339 time in UTC; null, empty and unparsable
// values are nil.
func (d *Document) Time(key string) *time.Time {
	return ParseTime(d.String(key))
}

// ParseTime parses the timestamps of the mirrors: RFC 3339, with or without
// seconds, in UTC.
func ParseTime(s string) *time.Time {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	for _, layout := range []string{time.RFC3339, "2006-01-02T15:04Z07:00", "2006-01-02"} {
		if t, err := time.Parse(layout, s); err == nil {
			t = t.UTC()
			return &t
		}
	}
	return nil
}

// Hash returns the version of a file: "sha256-" and the hex digest.
func Hash(raw []byte) string {
	sum := sha256.Sum256(raw)
	return "sha256-" + hex.EncodeToString(sum[:])
}

// normalizeMap converts the yaml.v2 maps with any keys into string-keyed
// maps, recursively.
func normalizeMap(m map[string]any) map[string]any {
	for k, v := range m {
		m[k] = normalizeValue(v)
	}
	return m
}

func normalizeValue(v any) any {
	switch v := v.(type) {
	case map[any]any:
		out := make(map[string]any, len(v))
		for k, item := range v {
			out[fmt.Sprint(k)] = normalizeValue(item)
		}
		return out
	case map[string]any:
		return normalizeMap(v)
	case []any:
		for i := range v {
			v[i] = normalizeValue(v[i])
		}
		return v
	default:
		return v
	}
}

// Stem is the file name without its extension: the code of an old-generation
// file and the name the wiki links use.
func Stem(path string) string {
	return strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
}

// WikiLink strips [[ and ]] from a link; other text stays as it is.
func WikiLink(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "[[") && strings.HasSuffix(s, "]]") {
		s = strings.TrimSpace(s[2 : len(s)-2])
		if i := strings.IndexAny(s, "|#"); i >= 0 {
			s = s[:i]
		}
	}
	return s
}
