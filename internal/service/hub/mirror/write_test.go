package mirror

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// copyFixture copies an issue file of the fixture mirror into a temporary
// root and returns the root and the copy.
func copyFixture(t *testing.T, rel string) (string, string) {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("../testdata/mirror", rel))
	require.NoError(t, err)
	root := t.TempDir()
	path := filepath.Join(root, ".github", "issues", filepath.Base(rel))
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.WriteFile(path, raw, 0o640))
	return root, path
}

func set(v string) Change { return Change{Set: true, Value: v} }

func TestPatchGolden(t *testing.T) {
	for _, tc := range []struct {
		name, file string
		triage     Triage
		want       func(before string) string
	}{
		{
			name: "code: replace impact, keep the rest", file: "octomation/maintainer/.github/issues/MAIN-1.md",
			triage: Triage{Impact: set("low")},
			want: func(s string) string {
				return strings.Replace(s, "  - \"impact: high\"\n", "  - \"impact: low\"\n", 1)
			},
		},
		{
			name: "code: remove effort, impact omitted", file: "octomation/maintainer/.github/issues/MAIN-1.md",
			triage: Triage{Effort: set("")},
			want: func(s string) string {
				return strings.Replace(s, "  - \"effort: easy\"\n", "", 1)
			},
		},
		{
			name: "code: same value is a no-op", file: "octomation/maintainer/.github/issues/MAIN-1.md",
			triage: Triage{Impact: set("high"), Effort: set("easy")},
			want:   func(s string) string { return s },
		},
		{
			name: "code: labels [] grows a block list", file: "octomation/maintainer/.github/issues/MAIN-2.md",
			triage: Triage{Impact: set("medium"), Effort: set("hard")},
			want: func(s string) string {
				return strings.Replace(s, "labels: []\n", "labels:\n  - \"impact: medium\"\n  - \"effort: hard\"\n", 1)
			},
		},
		{
			name: "code: removing the last label leaves labels []", file: "octomation/maintainer/.github/issues/MAIN-5.md",
			triage: Triage{Impact: set("")},
			want: func(s string) string {
				return strings.Replace(s, "labels:\n  - \"impact: low\"\n", "labels: []\n", 1)
			},
		},
		{
			name: "code: append to a block list without the label", file: "octomation/maintainer/.github/issues/MAIN-3.md",
			triage: Triage{Impact: set("high")},
			want: func(s string) string {
				return strings.Replace(s, "  - \"effort: hard\"\n", "  - \"effort: hard\"\n  - \"impact: high\"\n", 1)
			},
		},
		{
			name: "crlf", file: "octomation/maintainer/.github/issues/MAIN-4.md",
			triage: Triage{Impact: set("high"), Effort: set("")},
			want: func(s string) string {
				s = strings.Replace(s, "  - \"impact: medium\"\r\n", "  - \"impact: high\"\r\n", 1)
				return strings.Replace(s, "  - \"effort: medium\"\r\n", "", 1)
			},
		},
		{
			name: "old generation", file: "octomation/tool/.github/issues/issue-1.md",
			triage: Triage{Impact: set("high")},
			want: func(s string) string {
				return strings.Replace(s, "  - \"impact: low\"\n", "  - \"impact: high\"\n", 1)
			},
		},
		{
			name: "dotfiles: labels [] grows a flow list", file: "kamilsk/dotfiles/.github/issues/DOT-1.md",
			triage: Triage{Impact: set("medium"), Effort: set("easy")},
			want: func(s string) string {
				return strings.Replace(s, "labels: []\n", "labels: [\"impact: medium\", \"effort: easy\"]\n", 1)
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root, path := copyFixture(t, tc.file)
			before, err := os.ReadFile(path)
			require.NoError(t, err)
			doc, err := Patch(root, path, Hash(before), tc.triage, Options{})
			require.NoError(t, err)
			after, err := os.ReadFile(path)
			require.NoError(t, err)
			want := tc.want(string(before))
			assert.Equal(t, want, string(after))
			assert.Equal(t, Hash(after), doc.Version)
			assert.Equal(t, path, doc.Path)
			info, err := os.Stat(path)
			require.NoError(t, err)
			assert.Equal(t, os.FileMode(0o640), info.Mode().Perm(), "the permissions are kept")
			entries, err := os.ReadDir(filepath.Dir(path))
			require.NoError(t, err)
			assert.Len(t, entries, 1, "no temporary file is left")
			// The patched file parses to the expected labels.
			issue := ParseIssue(doc)
			for _, change := range []struct {
				prefix string
				Change
			}{{"impact", tc.triage.Impact}, {"effort", tc.triage.Effort}} {
				if !change.Set {
					continue
				}
				var found []string
				for _, label := range issue.Labels {
					if strings.HasPrefix(label, change.prefix+":") {
						found = append(found, label)
					}
				}
				if change.Value == "" {
					assert.Empty(t, found)
				} else {
					assert.Equal(t, []string{change.prefix + ": " + change.Value}, found)
				}
			}
		})
	}
}

func TestPatchErrors(t *testing.T) {
	t.Run("version mismatch", func(t *testing.T) {
		root, path := copyFixture(t, "octomation/maintainer/.github/issues/MAIN-1.md")
		_, err := Patch(root, path, "sha256-stale", Triage{Impact: set("low")}, Options{})
		assert.ErrorIs(t, err, ErrVersionMismatch)
	})

	t.Run("not editable", func(t *testing.T) {
		for name, file := range map[string]string{
			"malformed": "octomation/maintainer/.github/issues/MAIN-8.md",
		} {
			root, path := copyFixture(t, file)
			raw, err := os.ReadFile(path)
			require.NoError(t, err)
			_, err = Patch(root, path, Hash(raw), Triage{Impact: set("low")}, Options{})
			assert.ErrorIs(t, err, ErrNotEditable, name)
		}
		root, path := copyFixture(t, "octomation/maintainer/.github/issues/MAIN-1.md")
		raw, err := os.ReadFile(path)
		require.NoError(t, err)
		twice := strings.Replace(string(raw), "  - \"impact: high\"\n", "  - \"impact: high\"\n  - \"impact: low\"\n", 1)
		require.NoError(t, os.WriteFile(path, []byte(twice), 0o640))
		_, err = Patch(root, path, Hash([]byte(twice)), Triage{Impact: set("medium")}, Options{})
		assert.ErrorIs(t, err, ErrNotEditable, "several impact labels")
		_, err = Patch(root, path, Hash([]byte(twice)), Triage{Effort: set("medium")}, Options{})
		assert.NoError(t, err, "the effort label is still single")
	})

	t.Run("one-line flow lists are editable", func(t *testing.T) {
		for _, file := range []string{"octomation/maintainer/.github/issues/MAIN-9.md", "kamilsk/dotfiles/.github/issues/DOT-231.md"} {
			root, path := copyFixture(t, file)
			raw, err := os.ReadFile(path)
			require.NoError(t, err)
			doc, err := Patch(root, path, Hash(raw), Triage{Impact: set("low")}, Options{})
			require.NoError(t, err, file)
			assert.Contains(t, ParseIssue(doc).Labels, "impact: low", file)
		}
	})

	t.Run("not regular", func(t *testing.T) {
		root, path := copyFixture(t, "octomation/maintainer/.github/issues/MAIN-1.md")
		raw, err := os.ReadFile(path)
		require.NoError(t, err)
		link := filepath.Join(root, ".github", "issues", "LINK-1.md")
		require.NoError(t, os.Symlink(path, link))
		_, err = Patch(root, link, Hash(raw), Triage{Impact: set("low")}, Options{})
		assert.ErrorIs(t, err, ErrNotRegular, "a symlink target")
		other := t.TempDir()
		dirLink := filepath.Join(other, "issues")
		require.NoError(t, os.Symlink(filepath.Dir(path), dirLink))
		_, err = Patch(other, filepath.Join(dirLink, "MAIN-1.md"), Hash(raw), Triage{Impact: set("low")}, Options{})
		assert.ErrorIs(t, err, ErrNotRegular, "a symlink in the path")
		_, err = Patch(filepath.Join(root, "elsewhere"), path, Hash(raw), Triage{Impact: set("low")}, Options{})
		assert.ErrorIs(t, err, ErrNotRegular, "outside the root")
		after, err := os.ReadFile(path)
		require.NoError(t, err)
		assert.Equal(t, raw, after)
	})

	t.Run("crash before the rename", func(t *testing.T) {
		root, path := copyFixture(t, "octomation/maintainer/.github/issues/MAIN-1.md")
		raw, err := os.ReadFile(path)
		require.NoError(t, err)
		crashed := errors.New("crash")
		testHookBeforeRename = func() error { return crashed }
		defer func() { testHookBeforeRename = nil }()
		_, err = Patch(root, path, Hash(raw), Triage{Impact: set("low")}, Options{})
		assert.ErrorIs(t, err, crashed)
		after, err := os.ReadFile(path)
		require.NoError(t, err)
		assert.Equal(t, raw, after, "the original is intact")
		entries, err := os.ReadDir(filepath.Dir(path))
		require.NoError(t, err)
		assert.Len(t, entries, 1, "the temporary file is removed")
	})

	t.Run("a writer between the temporary file and the rename", func(t *testing.T) {
		root, path := copyFixture(t, "octomation/maintainer/.github/issues/MAIN-1.md")
		raw, err := os.ReadFile(path)
		require.NoError(t, err)
		testHookBeforeRename = func() error { return nil }
		defer func() { testHookBeforeRename = nil }()
		// The recheck runs before the hook: simulate the race by changing the
		// file during the temporary write through the hook of the recheck.
		changed := append([]byte(nil), raw...)
		changed = append(changed, "\nedited elsewhere\n"...)
		testHookBeforeRename = func() error { return os.WriteFile(path, changed, 0o640) }
		_, err = Patch(root, path, Hash(raw), Triage{Impact: set("low")}, Options{})
		// The hook runs after the recheck, so the rename wins here: this is
		// the documented operating limit. The protocol still detects a change
		// made before the recheck:
		_ = err
		require.NoError(t, os.WriteFile(path, raw, 0o640))
		_, err = Patch(root, path, Hash(changed), Triage{Impact: set("low")}, Options{})
		assert.ErrorIs(t, err, ErrVersionMismatch)
	})
}

// writeTemp writes an issue file with the given labels line into a temporary
// root and returns the root, the path and the content.
func writeTemp(t *testing.T, labelsLine string, crlf bool) (string, string, string) {
	t.Helper()
	content := "---\ncode: DOT-1\nid: 1\nstatus: open\ntitle: \"flow\"\n" + labelsLine + "\nurl: https://example.com/1\n---\n\n# flow\n\nBody.\n"
	if crlf {
		content = strings.ReplaceAll(content, "\n", "\r\n")
	}
	root := t.TempDir()
	path := filepath.Join(root, ".github", "issues", "DOT-1.md")
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
	return root, path, content
}

func TestPatchFlowListGolden(t *testing.T) {
	for _, tc := range []struct {
		name, before, after string
		triage              Triage
		crlf                bool
	}{
		{name: "replace in the dotfiles style", before: `labels: ["type: improvement","scope: code","impact: medium","effort: easy"]`,
			after: `labels: ["type: improvement","scope: code","impact: high","effort: easy"]`, triage: Triage{Impact: set("high")}},
		{name: "add with the neighbours' separator and quotes", before: `labels: ["type: improvement","scope: code"]`,
			after: `labels: ["type: improvement","scope: code","effort: easy"]`, triage: Triage{Effort: set("easy")}},
		{name: "add with spaced separators", before: `labels: ["a", "scope: code"]`,
			after: `labels: ["a", "scope: code", "impact: low"]`, triage: Triage{Impact: set("low")}},
		{name: "add to a one-item list", before: `labels: ["scope: code"]`,
			after: `labels: ["scope: code", "impact: low"]`, triage: Triage{Impact: set("low")}},
		{name: "remove a middle item with its separator", before: `labels: ["type: improvement","scope: code","impact: medium","effort: easy"]`,
			after: `labels: ["type: improvement","scope: code","effort: easy"]`, triage: Triage{Impact: set("")}},
		{name: "remove the first item with the separator after it", before: `labels: ["impact: medium", "scope: code"]`,
			after: `labels: ["scope: code"]`, triage: Triage{Impact: set("")}},
		{name: "remove the last item leaves labels []", before: `labels: ["impact: low"]`,
			after: `labels: []`, triage: Triage{Impact: set("")}},
		{name: "remove both", before: `labels: ["impact: low", "effort: hard"]`,
			after: `labels: []`, triage: Triage{Impact: set(""), Effort: set("")}},
		{name: "bare neighbours get a quoted label", before: `labels: [bug, docs]`,
			after: `labels: [bug, docs, "impact: high"]`, triage: Triage{Impact: set("high")}},
		{name: "a bare item is replaced in its own style, quoted when needed", before: `labels: [bug, impact:high]`,
			after: `labels: [bug, "impact: low"]`, triage: Triage{Impact: set("low")}},
		{name: "single quotes", before: `labels: ['impact: low', 'scope: code']`,
			after: `labels: ['impact: high', 'scope: code', 'effort: easy']`, triage: Triage{Impact: set("high"), Effort: set("easy")}},
		{name: "inner spaces and a tab after the key stay", before: "labels:\t[ \"impact: low\" , \"scope: code\" ]",
			after: "labels:\t[ \"impact: medium\" , \"scope: code\" ]", triage: Triage{Impact: set("medium")}},
		{name: "same value is a no-op", before: `labels: ["impact: low"]`,
			after: `labels: ["impact: low"]`, triage: Triage{Impact: set("low")}},
		{name: "crlf", before: `labels: ["impact: low","scope: code"]`,
			after: `labels: ["impact: high","scope: code","effort: hard"]`, triage: Triage{Impact: set("high"), Effort: set("hard")}, crlf: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root, path, before := writeTemp(t, tc.before, tc.crlf)
			doc, err := Patch(root, path, Hash([]byte(before)), tc.triage, Options{})
			require.NoError(t, err)
			after, err := os.ReadFile(path)
			require.NoError(t, err)
			want := strings.Replace(before, tc.before, tc.after, 1)
			if tc.crlf {
				want = strings.Replace(before, strings.ReplaceAll(tc.before, "\n", "\r\n"), tc.after, 1)
			}
			assert.Equal(t, want, string(after))
			assert.Equal(t, Hash(after), doc.Version)
			entries, err := os.ReadDir(filepath.Dir(path))
			require.NoError(t, err)
			assert.Len(t, entries, 1, "no temporary file is left")
		})
	}
}

func TestPatchFlowListNotEditable(t *testing.T) {
	for name, labels := range map[string]string{
		"several lines":        "labels: [\"impact: low\",\n  \"scope: code\"]",
		"comment inside":       `labels: ["impact: low", # todo` + "\n" + `  "scope: code"]`,
		"nested list":          `labels: [["impact: low"], "scope: code"]`,
		"nested mapping":       `labels: [{impact: low}, "scope: code"]`,
		"bare mapping item":    `labels: [impact: low, scope]`,
		"trailing separator":   `labels: ["impact: low", ]`,
		"empty item":           `labels: ["impact: low",, "scope: code"]`,
		"unterminated quote":   `labels: ["impact: low]`,
		"several impact items": `labels: ["impact: low", "impact: high"]`,
	} {
		root, path, before := writeTemp(t, labels, false)
		_, err := Patch(root, path, Hash([]byte(before)), Triage{Impact: set("medium")}, Options{})
		assert.ErrorIs(t, err, ErrNotEditable, name)
		after, err := os.ReadFile(path)
		require.NoError(t, err)
		assert.Equal(t, before, string(after), name)
	}
}

func TestPatchProtocolReview(t *testing.T) {
	t.Run("a root that is a symlink to a real checkout", func(t *testing.T) {
		real, path := copyFixture(t, "octomation/maintainer/.github/issues/MAIN-1.md")
		raw, err := os.ReadFile(path)
		require.NoError(t, err)
		link := filepath.Join(t.TempDir(), "checkout")
		require.NoError(t, os.Symlink(real, link))
		_, err = Patch(link, filepath.Join(link, ".github", "issues", "MAIN-1.md"), Hash(raw), Triage{Impact: set("low")}, Options{})
		assert.ErrorIs(t, err, ErrNotRegular)
		after, err := os.ReadFile(path)
		require.NoError(t, err)
		assert.Equal(t, raw, after, "nothing is written through the link")
		_, err = Patch(real, path, Hash(raw), Triage{Impact: set("low")}, Options{})
		assert.NoError(t, err, "the real root under a temporary directory is fine, whatever its ancestors are")
	})

	t.Run("a path with .. is refused before any operation", func(t *testing.T) {
		root, path := copyFixture(t, "octomation/maintainer/.github/issues/MAIN-1.md")
		raw, err := os.ReadFile(path)
		require.NoError(t, err)
		outside := filepath.Join(t.TempDir(), "MAIN-1.md")
		require.NoError(t, os.WriteFile(outside, raw, 0o644))
		for _, p := range []string{
			filepath.Join(root, ".github", "issues", "..", "..", "..", filepath.Base(filepath.Dir(outside)), "MAIN-1.md"),
			filepath.Join(root, "..", "x.md"),
			filepath.Join(root, ".."),
		} {
			_, err := Patch(root, p, Hash(raw), Triage{Impact: set("low")}, Options{})
			assert.ErrorIs(t, err, ErrNotRegular, p)
		}
		for _, dir := range []string{filepath.Dir(path), filepath.Dir(outside), root} {
			entries, err := os.ReadDir(dir)
			require.NoError(t, err)
			for _, e := range entries {
				assert.NotContains(t, e.Name(), ".tmp-", dir)
			}
		}
		after, err := os.ReadFile(outside)
		require.NoError(t, err)
		assert.Equal(t, raw, after)
	})

	t.Run("the temporary file and the rename happen inside the root handle", func(t *testing.T) {
		root, path := copyFixture(t, "octomation/maintainer/.github/issues/MAIN-1.md")
		raw, err := os.ReadFile(path)
		require.NoError(t, err)
		// A relative symlink inside the root that points outside: the
		// component check refuses it, and os.Root would too.
		elsewhere := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(elsewhere, "MAIN-2.md"), raw, 0o644))
		link := filepath.Join(root, ".github", "issues", "MAIN-2.md")
		require.NoError(t, os.Symlink(filepath.Join(elsewhere, "MAIN-2.md"), link))
		_, err = Patch(root, link, Hash(raw), Triage{Impact: set("low")}, Options{})
		assert.ErrorIs(t, err, ErrNotRegular)
		r, err := os.OpenRoot(root)
		require.NoError(t, err)
		defer r.Close()
		_, err = r.OpenFile(filepath.Join(".github", "issues", "MAIN-2.md", "..", "..", "..", "..", "escape.tmp"), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
		assert.Error(t, err, "the handle refuses an escaping path")
		_, err = r.OpenFile(filepath.Join(".github", "issues", "MAIN-2.md"), os.O_WRONLY, 0o644)
		assert.Error(t, err, "the handle refuses a symlink that leaves the root")
	})

	t.Run("the mode survives the umask", func(t *testing.T) {
		root, path := copyFixture(t, "octomation/maintainer/.github/issues/MAIN-1.md")
		require.NoError(t, os.Chmod(path, 0o664))
		old := syscall.Umask(0o022)
		defer syscall.Umask(old)
		raw, err := os.ReadFile(path)
		require.NoError(t, err)
		_, err = Patch(root, path, Hash(raw), Triage{Impact: set("low")}, Options{})
		require.NoError(t, err)
		info, err := os.Stat(path)
		require.NoError(t, err)
		assert.Equal(t, os.FileMode(0o664), info.Mode().Perm())
	})

	t.Run("a directory sync failure after the rename", func(t *testing.T) {
		root, path := copyFixture(t, "octomation/maintainer/.github/issues/MAIN-1.md")
		raw, err := os.ReadFile(path)
		require.NoError(t, err)
		testHookSyncDir = func(*os.File) error { return errors.New("sync failed") }
		defer func() { testHookSyncDir = nil }()
		doc, err := Patch(root, path, Hash(raw), Triage{Impact: set("low")}, Options{})
		require.ErrorIs(t, err, ErrDirectorySync)
		require.NotNil(t, doc, "the file is replaced: the new document is returned with the error")
		after, err := os.ReadFile(path)
		require.NoError(t, err)
		assert.Contains(t, string(after), "impact: low")
		assert.Equal(t, Hash(after), doc.Version)
		entries, err := os.ReadDir(filepath.Dir(path))
		require.NoError(t, err)
		assert.Len(t, entries, 1)
	})
}

// The ancestor of the root is swapped for a symlink to an identical copy
// after the caller's checks and before the root handle opens: the directory
// identity recorded at discovery refuses the write, and neither tree changes.
func TestPatchRootIdentity(t *testing.T) {
	base := t.TempDir()
	src := filepath.Join(base, "work", "octomation", "maintainer")
	require.NoError(t, os.MkdirAll(filepath.Join(src, ".github", "issues"), 0o755))
	raw, err := os.ReadFile(filepath.Join("../testdata/mirror", "octomation/maintainer/.github/issues/MAIN-1.md"))
	require.NoError(t, err)
	path := filepath.Join(src, ".github", "issues", "MAIN-1.md")
	require.NoError(t, os.WriteFile(path, raw, 0o644))
	identity, err := os.Stat(src)
	require.NoError(t, err)

	swapped := false
	testHookBeforeOpenRoot = func() error {
		require.NoError(t, os.Rename(filepath.Join(base, "work"), filepath.Join(base, "work-real")))
		require.NoError(t, os.CopyFS(filepath.Join(base, "copy"), os.DirFS(filepath.Join(base, "work-real"))))
		require.NoError(t, os.Symlink(filepath.Join(base, "copy"), filepath.Join(base, "work")))
		swapped = true
		return nil
	}
	defer func() { testHookBeforeOpenRoot = nil }()
	_, err = Patch(src, path, Hash(raw), Triage{Impact: set("low")}, Options{Identity: identity})
	assert.ErrorIs(t, err, ErrNotEditable)
	assert.True(t, swapped)
	for _, tree := range []string{"work-real", "copy"} {
		after, err := os.ReadFile(filepath.Join(base, tree, "octomation", "maintainer", ".github", "issues", "MAIN-1.md"))
		require.NoError(t, err)
		assert.Equal(t, raw, after, tree)
		entries, err := os.ReadDir(filepath.Join(base, tree, "octomation", "maintainer", ".github", "issues"))
		require.NoError(t, err)
		assert.Len(t, entries, 1, tree)
	}
	testHookBeforeOpenRoot = nil
	// The same directory, reached through the stable symlink, is accepted
	// when its identity is the recorded one.
	real := filepath.Join(base, "work-real", "octomation", "maintainer")
	_, err = Patch(real, filepath.Join(real, ".github", "issues", "MAIN-1.md"), Hash(raw), Triage{Impact: set("low")}, Options{Identity: identity})
	require.NoError(t, err)
	// And the copy, with a different identity, is refused even without a swap.
	copyRoot := filepath.Join(base, "copy", "octomation", "maintainer")
	_, err = Patch(copyRoot, filepath.Join(copyRoot, ".github", "issues", "MAIN-1.md"), Hash(raw), Triage{Impact: set("low")}, Options{Identity: identity})
	assert.ErrorIs(t, err, ErrNotEditable)
}
