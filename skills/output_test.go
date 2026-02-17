package skills_test

import (
	"bytes"
	"encoding/json"
	"io"
	"io/fs"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	fetchsvc "go.octolab.org/toolset/maintainer/internal/service/fetch"
	"go.octolab.org/toolset/maintainer/internal/service/status"
)

// TestOutput keeps the fields that references/output.md names in step with
// the JSON that fetch and status print.
func TestOutput(t *testing.T) {
	fields := documented(t)

	// A plan whose actions fill every optional field.
	plan := fetchsvc.Plan{
		ID:          "plan",
		GeneratedAt: time.Unix(0, 0),
		Discoveries: []fetchsvc.DiscoverySummary{{Profile: "primary", Count: 2}},
		Actions: []fetchsvc.Action{
			{Kind: fetchsvc.KindMove, ID: 1, Owner: "acme", Name: "new", FromName: "acme/old",
				FromPath: "/old", ToPath: "/new", Worktrees: []string{"/wt"}, Flag: "unverified", Reason: "renamed"},
			{Kind: fetchsvc.KindOrphan, ID: 2, Owner: "acme", Name: "dup", Path: "/dup",
				OrphanReason: "duplicate", ActivePath: "/active"},
		},
	}
	var out bytes.Buffer
	require.NoError(t, fetchsvc.NewReporter(&out, io.Discard, fetchsvc.FormatJSON, 0, false).Render(plan, false))
	var doc map[string]any
	require.NoError(t, json.Unmarshal(out.Bytes(), &doc))
	have := map[string]bool{}
	keys(doc, have)
	for _, name := range fields["fetch --format=json"] {
		assert.True(t, have[name], "references/output.md names %s, which fetch --format=json lacks", name)
	}

	// Every field of a status row, and nothing else, is documented.
	var row []string
	for _, field := range reflect.VisibleFields(reflect.TypeOf(status.Row{})) {
		if name, _, _ := strings.Cut(field.Tag.Get("json"), ","); name != "" && name != "-" {
			row = append(row, name)
		}
	}
	assert.ElementsMatch(t, row, fields["status --format=json"])
}

// documented returns, per section of references/output.md, the backquoted
// names in the first column of its tables of fields and of kinds, which the
// summary of a plan counts by name.
func documented(t *testing.T) map[string][]string {
	t.Helper()
	data, err := fs.ReadFile(maintainer, "references/output.md")
	require.NoError(t, err)
	names := regexp.MustCompile("`([a-z_]+)`")
	fields := map[string][]string{}
	var section string
	var table bool
	for _, line := range strings.Split(string(data), "\n") {
		switch {
		case strings.HasPrefix(line, "## "):
			section = strings.TrimPrefix(line, "## ")
		case strings.HasPrefix(line, "| Field |"), strings.HasPrefix(line, "| Kind |"):
			table = true
		case !strings.HasPrefix(line, "|"):
			table = false
		case table:
			cell := strings.Split(line, "|")[1]
			for _, m := range names.FindAllStringSubmatch(cell, -1) {
				fields[section] = append(fields[section], m[1])
			}
		}
	}
	require.NotEmpty(t, fields["fetch --format=json"])
	require.NotEmpty(t, fields["status --format=json"])
	return fields
}

// keys collects the object keys at every depth of a decoded JSON value.
func keys(v any, into map[string]bool) {
	switch v := v.(type) {
	case map[string]any:
		for k, e := range v {
			into[k] = true
			keys(e, into)
		}
	case []any:
		for _, e := range v {
			keys(e, into)
		}
	}
}
