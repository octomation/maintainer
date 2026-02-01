package config

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var update = flag.Bool("update", false, "rewrite the fixtures of the template and the guide examples")

// The TOML and YAML spellings of every key are the same.
func TestFetchKeysMatchInTOMLAndYAML(t *testing.T) {
	for typ := range configTypes(reflect.TypeFor[Fetch]()) {
		for i := range typ.NumField() {
			field := typ.Field(i)
			assert.Equalf(t, field.Tag.Get("toml"), field.Tag.Get("yaml"), "%s.%s: TOML and YAML keys differ", typ.Name(), field.Name)
		}
	}
}

func configTypes(typ reflect.Type) map[reflect.Type]bool {
	out := map[reflect.Type]bool{}
	var walk func(reflect.Type)
	walk = func(typ reflect.Type) {
		for typ.Kind() == reflect.Pointer || typ.Kind() == reflect.Slice || typ.Kind() == reflect.Map {
			typ = typ.Elem()
		}
		if typ.Kind() != reflect.Struct || out[typ] {
			return
		}
		out[typ] = true
		for _, field := range structKeys(typ) {
			walk(field.Type)
		}
	}
	walk(typ)
	return out
}

// The fixtures are the ones `make config-vet` checks against fetch.cue, so
// the loader and the schema accept and reject the same files: at the
// boundaries of the values, unknown keys and, among the valid ones, the
// template of `fetch config init` and the config examples of the guides. The
// loader-only checks are listed in fetch.cue.
func TestFetchSchemaFixtures(t *testing.T) {
	for dir, valid := range map[string]bool{"valid": true, "invalid": false} {
		var paths []string
		for _, ext := range []string{"toml", "yaml"} {
			found, err := filepath.Glob(filepath.Join("testdata", dir, "*."+ext))
			require.NoError(t, err)
			require.NotEmptyf(t, found, "config-vet needs a %s fixture in testdata/%s", ext, dir)
			paths = append(paths, found...)
		}
		for _, path := range paths {
			cnf, err := LoadFetch(afero.NewOsFs(), path)
			if err == nil {
				err = cnf.Validate()
			}
			assert.Equal(t, valid, err == nil, "%s: %v", path, err)
		}
	}
}

// The fixtures of the template and of the guide examples, with their
// commented-out keys enabled as well, are current: `go test ./internal/config
// -run TestFetchExampleFixtures -update` rewrites them.
func TestFetchExampleFixtures(t *testing.T) {
	examples := guideExamples(t, "fetch")
	require.GreaterOrEqual(t, len(examples), 4, "the guides lost their config examples")
	examples = append(examples, example{name: "template", file: "fetch.toml", body: FetchConfigTemplate})
	checkExampleFixtures(t, examples, filepath.Join("testdata", "valid"), []string{"guide-*", "template*"}, "TestFetchExampleFixtures")
}

// The template the service writes and the service.toml examples of the
// guides (fences with filename="service.toml"), with their commented-out keys
// enabled as well, are fixtures that `make config-vet` checks against
// service.cue: `go test ./internal/config -run TestServiceExampleFixtures
// -update` rewrites them.
func TestServiceExampleFixtures(t *testing.T) {
	examples := guideExamples(t, "service")
	require.GreaterOrEqual(t, len(examples), 2, "the service guide lost its service.toml examples")
	examples = append(examples, example{name: "template", file: "service.toml", body: string(ServiceTemplate)})
	checkExampleFixtures(t, examples, filepath.Join("testdata", "service", "valid"), []string{"guide-*", "template*"}, "TestServiceExampleFixtures")
}

func checkExampleFixtures(t *testing.T, examples []example, dir string, patterns []string, test string) {
	t.Helper()
	want := map[string]string{}
	for _, ex := range examples {
		ext := filepath.Ext(ex.file)
		want[ex.name+ext] = ex.body
		if enabled := uncomment(ex.body); enabled != ex.body {
			want[ex.name+"-uncommented"+ext] = enabled
		}
	}
	generated := func() map[string]string {
		var paths []string
		for _, pattern := range patterns {
			found, err := filepath.Glob(filepath.Join(dir, pattern))
			require.NoError(t, err)
			paths = append(paths, found...)
		}
		out := map[string]string{}
		for _, path := range paths {
			raw, err := os.ReadFile(path)
			require.NoError(t, err)
			out[filepath.Base(path)] = string(raw)
		}
		return out
	}
	got := generated()
	if *update {
		for name := range got {
			require.NoError(t, os.Remove(filepath.Join(dir, name)))
		}
		for name, body := range want {
			require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644))
		}
		got = generated()
	}
	assert.Equal(t, want, got, "run go test ./internal/config -run "+test+" -update")
}

type example struct{ name, file, body string }

var (
	fence   = regexp.MustCompile("^```(\\w*)(?: filename=\"((?:fetch|service)\\.toml)\")?")
	heredoc = regexp.MustCompile(`cat > (?:\S*/)?((?:fetch|service)\.(?:toml|yaml)) <<'?EOF'?`)
	setting = regexp.MustCompile(`^(\s*)# ?(\[\[?[a-z_]+|[a-z_]+\s*=)`)
)

// guideExamples extracts every TOML block (a fetch config) and every fetch or
// service config written by a shell heredoc from the guides, named after the
// page and their order on it, and keeps those of kind ("fetch" or "service");
// the release notes are frozen and skipped.
func guideExamples(t *testing.T, kind string) []example {
	var kept []example
	for _, ex := range allGuideExamples(t) {
		if strings.HasPrefix(ex.file, kind+".") {
			kept = append(kept, ex)
		}
	}
	return kept
}

func allGuideExamples(t *testing.T) []example {
	t.Helper()
	pages, err := filepath.Glob(filepath.Join("..", "..", "docs", "content", "*.md"))
	require.NoError(t, err)
	require.NotEmpty(t, pages)
	var out []example
	for _, page := range pages {
		raw, err := os.ReadFile(page)
		require.NoError(t, err)
		var (
			lang, file string
			count      int
			body       bytes.Buffer
			inFence    bool
			inHeredoc  bool
		)
		name := func() string {
			count++
			return fmt.Sprintf("guide-%s-%d", strings.TrimSuffix(filepath.Base(page), ".md"), count)
		}
		scanner := bufio.NewScanner(bytes.NewReader(raw))
		for scanner.Scan() {
			line := scanner.Text()
			switch {
			case inHeredoc && line == "EOF":
				out = append(out, example{name: name(), file: file, body: body.String()})
				inHeredoc = false
			case inHeredoc:
				body.WriteString(line + "\n")
			case !inFence && fence.MatchString(line):
				m := fence.FindStringSubmatch(line)
				inFence, lang, file = true, m[1], m[2]
				body.Reset()
			case inFence && strings.HasPrefix(line, "```"):
				if lang == "toml" {
					// A fence without a filename shows the fetch config.
					if file == "" {
						file = "fetch.toml"
					}
					out = append(out, example{name: name(), file: file, body: body.String()})
				}
				inFence = false
			case inFence && heredoc.MatchString(line):
				inHeredoc, file = true, heredoc.FindStringSubmatch(line)[1]
				body.Reset()
			case inFence:
				body.WriteString(line + "\n")
			}
		}
		require.NoError(t, scanner.Err())
	}
	return out
}

// uncomment enables the settings an example shows as comments, such as
// "# [[repos]]" or "# state_file = …", so they are checked too.
func uncomment(body string) string {
	lines := strings.Split(body, "\n")
	for i, line := range lines {
		if m := setting.FindStringSubmatchIndex(line); m != nil {
			lines[i] = line[m[2]:m[3]] + line[m[4]:]
		}
	}
	return strings.Join(lines, "\n")
}
