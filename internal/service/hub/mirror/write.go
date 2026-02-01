package mirror

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"gopkg.in/yaml.v2"
)

// Errors of a label patch, mapped to HTTP statuses by the API.
var (
	// ErrVersionMismatch reports a file whose content is not the expected version.
	ErrVersionMismatch = errors.New("the file changed")
	// ErrNotEditable reports a frontmatter the patch cannot edit in place.
	ErrNotEditable = errors.New("the labels cannot be edited in place")
	// ErrNotRegular reports a target that is not a regular file under the root
	// without a symlink in its path.
	ErrNotRegular = errors.New("not a regular file under the mirror root")
	// ErrDirectorySync reports a write whose file was replaced but whose
	// directory could not be synced afterwards: the content is in place.
	ErrDirectorySync = errors.New("the file was replaced but the directory sync failed")
)

// Change is one triage label edit: unset leaves the label, an empty Value
// removes it, another value replaces it.
type Change struct {
	Set   bool
	Value string
}

// Triage is the patch of the two triage labels.
type Triage struct {
	Impact Change
	Effort Change
}

// Empty reports a patch that changes nothing.
func (t Triage) Empty() bool { return !t.Impact.Set && !t.Effort.Set }

// writeMu serializes every write of the process: the hash check and the
// rename do not protect against a second writer of this service.
var writeMu sync.Mutex

// testHookBeforeRename lets a test crash the protocol between the temporary
// file and the rename.
var testHookBeforeRename func() error

// testHookBeforeOpenRoot lets a test change the tree between the caller's
// checks and the opening of the root handle.
var testHookBeforeOpenRoot func() error

// Options of a write.
type Options struct {
	// Identity is the root directory as the caller saw it at discovery: the
	// directory the handle opens must be the same file, so an ancestor
	// swapped for a symlink to a copy between the caller's checks and the
	// open redirects nothing. Nil skips the check.
	Identity os.FileInfo
}

var (
	labelsLine = regexp.MustCompile(`^(labels:)([ \t]*)(.*?)([ \t]*)$`)
	labelItem  = regexp.MustCompile(`^([ \t]+)-[ \t]+(.*?)[ \t]*$`)
)

// Patch applies a triage to the issue file at path, a regular file under
// mirrorRoot with no symlink in its path, whose current content must hash
// to version. Only the `impact: *` and `effort: *` entries of the labels
// change; every other byte stays. Every file operation goes through a
// directory handle of the root (os.Root), so a pathname that escapes the
// root through a symlink or .. is refused by the kernel, not by a check on
// the name. The write goes to an exclusively created temporary file next to
// the target with the same permissions, is fsynced, the original is hashed
// again, the file is renamed over it and the directory is fsynced. It
// returns the new document.
func Patch(mirrorRoot, path, version string, triage Triage, opts Options) (*Document, error) {
	writeMu.Lock()
	defer writeMu.Unlock()
	rel, err := relativeTarget(mirrorRoot, path)
	if err != nil {
		return nil, err
	}
	// The handle would follow a root that is itself a symlink: refuse it by
	// name, as the index refuses a root whose real path changed.
	if info, err := os.Lstat(mirrorRoot); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrNotRegular, err)
	} else if info.Mode()&os.ModeSymlink != 0 {
		return nil, fmt.Errorf("%w: the mirror root %s is a symlink", ErrNotRegular, mirrorRoot)
	}
	if testHookBeforeOpenRoot != nil {
		if err := testHookBeforeOpenRoot(); err != nil {
			return nil, err
		}
	}
	root, err := os.OpenRoot(mirrorRoot)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrNotRegular, err)
	}
	defer root.Close()
	if opts.Identity != nil {
		opened, err := root.Stat(".")
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrNotRegular, err)
		}
		if !os.SameFile(opts.Identity, opened) {
			return nil, fmt.Errorf("%w: the mirror root %s is not the directory seen at discovery", ErrNotEditable, mirrorRoot)
		}
	}
	if err := checkRegular(root, rel); err != nil {
		return nil, err
	}
	raw, err := root.ReadFile(rel)
	if err != nil {
		return nil, err
	}
	if Hash(raw) != version {
		return nil, ErrVersionMismatch
	}
	doc, err := Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrNotEditable, err)
	}
	patched, err := patchLabels(doc, triage)
	if err != nil {
		return nil, err
	}
	if bytes.Equal(patched, raw) {
		doc.Path = path
		return doc, nil
	}
	next, err := Parse(patched)
	if err != nil {
		return nil, fmt.Errorf("%w: the patched frontmatter does not parse: %v", ErrNotEditable, err)
	}
	err = writeAtomically(root, rel, raw, version, patched)
	if err != nil && !errors.Is(err, ErrDirectorySync) {
		return nil, err
	}
	next.Path = path
	return next, err
}

// patchLabels returns the patched bytes of doc. The labels are a block
// list, `labels: []`, or a flow list on one line whose items are plain
// scalars; only the triage items change and everything else stays.
func patchLabels(doc *Document, triage Triage) ([]byte, error) {
	if triage.Empty() {
		return doc.Raw, nil
	}
	nl := "\n"
	if doc.CRLF {
		nl = "\r\n"
	}
	front := string(doc.Frontmatter())
	lines := strings.Split(front, nl)
	start := -1
	for i, line := range lines {
		if labelsLine.MatchString(line) {
			if start >= 0 {
				return nil, fmt.Errorf("%w: several labels keys", ErrNotEditable)
			}
			start = i
		}
	}
	if start < 0 {
		return nil, fmt.Errorf("%w: no labels key", ErrNotEditable)
	}
	key := labelsLine.FindStringSubmatch(lines[start])
	inline := key[3]
	var list labelList
	end := start + 1
	switch {
	case inline == "[]" && doc.Has("status") && !doc.Has("state"):
		// The dotfiles generation writes flow lists: an empty one grows as one.
		list.kind = flowList
	case inline == "" || inline == "[]":
		list.kind = blockList
		list.indent = "  "
		for end < len(lines) {
			m := labelItem.FindStringSubmatch(lines[end])
			if m == nil {
				break
			}
			if inline == "[]" {
				return nil, fmt.Errorf("%w: labels: [] followed by items", ErrNotEditable)
			}
			list.indent = m[1]
			list.items = append(list.items, labelEntry{text: lines[end], value: unquote(m[2]), quote: quoteOf(m[2])})
			end++
		}
	case strings.HasPrefix(inline, "[") && strings.HasSuffix(inline, "]"):
		list.kind = flowList
		if err := list.parseFlow(inline[1 : len(inline)-1]); err != nil {
			return nil, err
		}
	case strings.HasPrefix(inline, "["):
		return nil, fmt.Errorf("%w: labels is a flow list over several lines", ErrNotEditable)
	default:
		return nil, fmt.Errorf("%w: labels is not a list", ErrNotEditable)
	}
	values := list.values()
	if parsed := doc.Strings("labels"); !equalStrings(parsed, values) {
		return nil, fmt.Errorf("%w: the labels list does not match the parsed list", ErrNotEditable)
	}
	changed := false
	for _, change := range []struct {
		prefix string
		Change
	}{{"impact", triage.Impact}, {"effort", triage.Effort}} {
		if !change.Set {
			continue
		}
		var found []int
		for i, v := range values {
			if strings.HasPrefix(v, change.prefix+":") {
				found = append(found, i)
			}
		}
		if len(found) > 1 {
			return nil, fmt.Errorf("%w: several %s labels", ErrNotEditable, change.prefix)
		}
		label := change.prefix + ": " + change.Value
		switch {
		case change.Value == "" && len(found) == 1:
			list.remove(found[0])
			changed = true
		case change.Value == "":
		case len(found) == 1:
			if values[found[0]] == label {
				continue
			}
			list.replace(found[0], label)
			changed = true
		default:
			list.append(label)
			changed = true
		}
		values = list.values()
	}
	if !changed {
		return doc.Raw, nil
	}
	// Rebuild the frontmatter: the key line, the items, the rest.
	var out []string
	out = append(out, lines[:start]...)
	out = append(out, list.render(key)...)
	out = append(out, lines[end:]...)
	patchedFront := strings.Join(out, nl)
	var check struct {
		Labels []string `yaml:"labels"`
	}
	if err := yaml.Unmarshal([]byte(patchedFront), &check); err != nil || !equalStrings(check.Labels, values) {
		return nil, fmt.Errorf("%w: the patched labels do not parse back", ErrNotEditable)
	}
	patched := make([]byte, 0, len(doc.Raw)+64)
	patched = append(patched, doc.Raw[:doc.frontStart]...)
	patched = append(patched, patchedFront...)
	patched = append(patched, doc.Raw[doc.frontEnd:]...)
	return patched, nil
}

type listKind int

const (
	blockList listKind = iota
	flowList
)

// labelEntry is one label as written: its text (a whole line of a block
// list, a scalar of a flow list), its value and its quoting.
type labelEntry struct {
	text  string
	value string
	quote byte // '"', '\'' or 0 for bare
}

// labelList is the labels list as written, with what the renderer needs to
// put it back byte for byte around the changed items.
type labelList struct {
	kind   listKind
	indent string       // block: the indentation of the items
	items  []labelEntry // the items in order
	seps   []string     // flow: the separator after each item but the last
	lead   string       // flow: the whitespace after [
	trail  string       // flow: the whitespace before ]
}

// parseFlow reads the inside of a one-line flow list: plain scalars, quoted
// or bare, separated by commas. Anything else is not editable in place.
func (l *labelList) parseFlow(inner string) error {
	i := 0
	l.lead = leadingSpace(inner)
	i += len(l.lead)
	if i == len(inner) {
		return nil
	}
	for {
		var text string
		switch {
		case i < len(inner) && (inner[i] == '"' || inner[i] == '\''):
			q := inner[i]
			j := i + 1
			for j < len(inner) && inner[j] != q {
				if q == '"' && inner[j] == '\\' {
					j++
				}
				j++
			}
			if j >= len(inner) {
				return fmt.Errorf("%w: unterminated quote in the labels flow list", ErrNotEditable)
			}
			text = inner[i : j+1]
			i = j + 1
		default:
			j := i
			for j < len(inner) && inner[j] != ',' {
				j++
			}
			text = strings.TrimRight(inner[i:j], " \t")
			if text == "" || strings.ContainsAny(text, "[]{}#\"'") || strings.Contains(text, ": ") || strings.HasPrefix(text, "- ") {
				return fmt.Errorf("%w: the labels flow list has an item that is not a plain scalar", ErrNotEditable)
			}
			i += len(text)
		}
		l.items = append(l.items, labelEntry{text: text, value: unquote(text), quote: quoteOf(text)})
		// The separator: spaces, a comma, spaces; or the end.
		j := i
		for j < len(inner) && (inner[j] == ' ' || inner[j] == '\t') {
			j++
		}
		if j == len(inner) {
			l.trail = inner[i:]
			return nil
		}
		if inner[j] != ',' {
			return fmt.Errorf("%w: the labels flow list has an item that is not a plain scalar", ErrNotEditable)
		}
		j++
		for j < len(inner) && (inner[j] == ' ' || inner[j] == '\t') {
			j++
		}
		if j == len(inner) {
			return fmt.Errorf("%w: the labels flow list ends with a separator", ErrNotEditable)
		}
		l.seps = append(l.seps, inner[i:j])
		i = j
	}
}

func (l *labelList) values() []string {
	out := make([]string, 0, len(l.items))
	for _, it := range l.items {
		out = append(out, it.value)
	}
	return out
}

// remove drops an item with its separator: the one before it, or the one
// after the first item.
func (l *labelList) remove(i int) {
	l.items = append(l.items[:i], l.items[i+1:]...)
	if l.kind == flowList && len(l.seps) > 0 {
		at := i - 1
		if at < 0 {
			at = 0
		}
		l.seps = append(l.seps[:at], l.seps[at+1:]...)
	}
}

// replace sets a new label on an item, keeping its quoting style.
func (l *labelList) replace(i int, label string) {
	l.items[i].value = label
	l.items[i].text = l.text(label, l.items[i].quote)
}

// append adds a label in the quoting style of its neighbours and with their
// separator, ", " for a list without one.
func (l *labelList) append(label string) {
	var quote byte = '"'
	if len(l.items) > 0 {
		quote = l.items[len(l.items)-1].quote
	}
	if l.kind == flowList && len(l.items) > 0 {
		sep := ", "
		if len(l.seps) > 0 {
			sep = l.seps[0]
		}
		l.seps = append(l.seps, sep)
	}
	l.items = append(l.items, labelEntry{text: l.text(label, quote), value: label, quote: quote})
}

// text renders a label in a quoting style; a bare label that YAML would not
// read as a plain scalar, such as one with ": ", is double-quoted anyway.
func (l *labelList) text(label string, quote byte) string {
	var scalar string
	switch {
	case quote == '\'' && !strings.Contains(label, "'"):
		scalar = "'" + label + "'"
	case quote == 0 && !strings.ContainsAny(label, "[]{}#\"',") && !strings.Contains(label, ": ") && !strings.HasSuffix(label, ":"):
		scalar = label
	default:
		scalar = quoteLabel(label)
	}
	if l.kind == blockList {
		return l.indent + "- " + scalar
	}
	return scalar
}

// render returns the lines of the list: the key line and, for a block list,
// its items.
func (l *labelList) render(key []string) []string {
	if len(l.items) == 0 {
		space := key[2]
		if space == "" {
			space = " "
		}
		return []string{key[1] + space + "[]" + key[4]}
	}
	if l.kind == blockList {
		out := []string{key[1]}
		for _, it := range l.items {
			out = append(out, it.text)
		}
		return out
	}
	var b strings.Builder
	b.WriteString(key[1])
	b.WriteString(key[2])
	b.WriteByte('[')
	b.WriteString(l.lead)
	for i, it := range l.items {
		if i > 0 {
			b.WriteString(l.seps[i-1])
		}
		b.WriteString(it.text)
	}
	b.WriteString(l.trail)
	b.WriteByte(']')
	b.WriteString(key[4])
	return []string{b.String()}
}

func leadingSpace(s string) string {
	return s[:len(s)-len(strings.TrimLeft(s, " \t"))]
}

func quoteOf(s string) byte {
	if len(s) >= 2 && (s[0] == '"' && s[len(s)-1] == '"' || s[0] == '\'' && s[len(s)-1] == '\'') {
		return s[0]
	}
	return 0
}

func unquote(s string) string {
	if len(s) >= 2 && (s[0] == '"' && s[len(s)-1] == '"' || s[0] == '\'' && s[len(s)-1] == '\'') {
		var v string
		if err := yaml.Unmarshal([]byte(s), &v); err == nil {
			return v
		}
	}
	return s
}

func quoteLabel(s string) string {
	return `"` + strings.ReplaceAll(strings.ReplaceAll(s, `\`, `\\`), `"`, `\"`) + `"`
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// relativeTarget returns the path of the target relative to the mirror
// root; a target outside the root, or the root itself, is refused.
func relativeTarget(mirrorRoot, path string) (string, error) {
	mirrorRoot, path = filepath.Clean(mirrorRoot), filepath.Clean(path)
	rel, err := filepath.Rel(mirrorRoot, path)
	if err != nil || rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", ErrNotRegular
	}
	return rel, nil
}

// checkRegular rejects a symlink anywhere in the path from the root down to
// the file, and anything but a regular file. The root itself and its
// ancestors are what the handle resolves: the index compares the real path
// of the root with the one it saw at discovery.
func checkRegular(root *os.Root, rel string) error {
	current := ""
	for _, part := range strings.Split(rel, string(filepath.Separator)) {
		current = filepath.Join(current, part)
		info, err := root.Lstat(current)
		if err != nil {
			return fmt.Errorf("%w: %v", ErrNotRegular, err)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("%w: %s is a symlink", ErrNotRegular, filepath.Join(root.Name(), current))
		}
	}
	info, err := root.Lstat(rel)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrNotRegular, err)
	}
	if !info.Mode().IsRegular() {
		return ErrNotRegular
	}
	return nil
}

// writeAtomically writes content over rel through a temporary file, every
// step through the root handle.
func writeAtomically(root *os.Root, rel string, original []byte, version string, content []byte) (err error) {
	info, err := root.Stat(rel)
	if err != nil {
		return err
	}
	dir := filepath.Dir(rel)
	var suffix [8]byte
	if _, err := rand.Read(suffix[:]); err != nil {
		return err
	}
	tmpRel := filepath.Join(dir, "."+filepath.Base(rel)+".tmp-"+hex.EncodeToString(suffix[:]))
	tmp, err := root.OpenFile(tmpRel, os.O_WRONLY|os.O_CREATE|os.O_EXCL, info.Mode().Perm())
	if err != nil {
		return err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tmp.Close()
			_ = root.Remove(tmpRel)
		}
	}()
	// The umask applies to the creation: restore the original mode.
	if err := tmp.Chmod(info.Mode().Perm()); err != nil {
		return err
	}
	if _, err := tmp.Write(content); err != nil {
		return err
	}
	if err := tmp.Sync(); err != nil {
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	// The original may have changed while the temporary file was written.
	again, err := root.ReadFile(rel)
	if err != nil {
		return err
	}
	if Hash(again) != version || !bytes.Equal(again, original) {
		return ErrVersionMismatch
	}
	if testHookBeforeRename != nil {
		if err := testHookBeforeRename(); err != nil {
			return err
		}
	}
	// The directory is opened before the rename: if it cannot be, nothing
	// is replaced.
	d, err := root.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()
	if err := root.Rename(tmpRel, rel); err != nil {
		return err
	}
	committed = true
	if err := syncDir(d); err != nil {
		return fmt.Errorf("%w: %v", ErrDirectorySync, err)
	}
	return nil
}

// testHookSyncDir lets a test fail the directory sync after the rename.
var testHookSyncDir func(*os.File) error

func syncDir(d *os.File) error {
	if testHookSyncDir != nil {
		return testHookSyncDir(d)
	}
	return d.Sync()
}
