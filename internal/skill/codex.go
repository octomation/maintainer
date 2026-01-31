package skill

import (
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

// The rules below mirror Codex 0.163.0, codex-rs/core-plugin-common:
// plugin_id.rs for plugin keys and installed.rs for the active version.

// codexLocalVersion is the cache directory Codex prefers to any version.
const codexLocalVersion = "local"

type codexPlugin struct {
	ID          string
	InstallPath string
	Ambiguous   []string // the version directories Codex picks among by chance
}

// codexPlugins reads the enabled plugins named like the skill from Codex's
// config.toml, read-only. Codex lists an installed plugin there as
// [plugins."<name>@<marketplace>"] and unpacks it into
// plugins/cache/<marketplace>/<name>/<version>; the copy listed is the
// version Codex loads, see codexActiveVersion. A plugin table without
// enabled counts as enabled; values under [plugins] that are not tables are
// skipped.
func codexPlugins(codexHome, name string) []codexPlugin {
	data, err := os.ReadFile(filepath.Join(codexHome, "config.toml"))
	if err != nil {
		return nil
	}
	var config struct {
		Plugins map[string]any `toml:"plugins"`
	}
	if toml.Unmarshal(data, &config) != nil {
		return nil
	}
	var list []codexPlugin
	for id, value := range config.Plugins {
		table, ok := value.(map[string]any)
		if !ok {
			continue
		}
		if enabled, ok := table["enabled"].(bool); ok && !enabled {
			continue
		}
		plugin, marketplace, ok := parseCodexPluginID(id)
		if !ok || plugin != name {
			continue
		}
		base := filepath.Join(codexHome, "plugins", "cache", marketplace, plugin)
		if version, ambiguous := codexActiveVersion(base); version != "" {
			list = append(list, codexPlugin{ID: id, InstallPath: filepath.Join(base, version), Ambiguous: ambiguous})
		}
	}
	sort.Slice(list, func(i, j int) bool { return list[i].ID < list[j].ID })
	return list
}

// parseCodexPluginID splits <plugin>@<marketplace> at the last @ and
// validates both as path segments, as PluginId::parse does.
func parseCodexPluginID(id string) (plugin, marketplace string, ok bool) {
	at := strings.LastIndex(id, "@")
	if at < 0 {
		return "", "", false
	}
	plugin, marketplace = id[:at], id[at+1:]
	if !validCodexSegment(plugin, true) || !validCodexSegment(marketplace, false) {
		return "", "", false
	}
	return plugin, marketplace, true
}

// validCodexSegment is validate_plugin_segment: ASCII letters, digits, _ and
// -, and for a plugin name dots that separate non-empty parts.
func validCodexSegment(segment string, dots bool) bool {
	if segment == "" {
		return false
	}
	if dots && (strings.HasPrefix(segment, ".") || strings.HasSuffix(segment, ".") || strings.Contains(segment, "..")) {
		return false
	}
	for _, r := range segment {
		if isASCIIAlnum(r) || r == '-' || r == '_' || dots && r == '.' {
			continue
		}
		return false
	}
	return true
}

// codexActiveVersion is active_plugin_version: among the directories with a
// valid version name, local wins, otherwise the greatest by
// compareCodexVersions. It returns "" when there is none.
//
// When semantic and other names are mixed, Codex's comparison is not
// transitive (1.9.0 < 1.10.0 < 1.11x < 1.9.0), and what it picks depends on
// the order on disk. The pick is then still returned, together with all the
// names, as ambiguous.
func codexActiveVersion(base string) (string, []string) {
	entries, err := os.ReadDir(base)
	if err != nil {
		return "", nil
	}
	var (
		best            string
		names           []string
		semantic, other bool
	)
	for _, e := range entries {
		name := e.Name()
		if !e.IsDir() || !validCodexVersion(name) {
			continue
		}
		if name == codexLocalVersion {
			return name, nil
		}
		names = append(names, name)
		if _, ok := parseSemver(name); ok {
			semantic = true
		} else {
			other = true
		}
		if best == "" || compareCodexVersions(name, best) > 0 {
			best = name
		}
	}
	if semantic && other {
		sort.Strings(names)
		return best, names
	}
	return best, nil
}

// validCodexVersion is validate_plugin_version_segment.
func validCodexVersion(version string) bool {
	if version == "" || version == "." || version == ".." {
		return false
	}
	for _, r := range version {
		if !isASCIIAlnum(r) && !strings.ContainsRune("-_.+", r) {
			return false
		}
	}
	return true
}

// compareCodexVersions is compare_plugin_versions: semantic versions compare
// as the Rust semver crate orders them, build metadata included; when either
// is not one, the names compare as strings.
func compareCodexVersions(a, b string) int {
	va, okA := parseSemver(a)
	vb, okB := parseSemver(b)
	if !okA || !okB {
		return strings.Compare(a, b)
	}
	for i := range va.core {
		if va.core[i] != vb.core[i] {
			if va.core[i] < vb.core[i] {
				return -1
			}
			return 1
		}
	}
	if c := comparePrerelease(va.pre, vb.pre); c != 0 {
		return c
	}
	return compareBuild(va.build, vb.build)
}

type codexSemver struct {
	core       [3]uint64
	pre, build string
}

// parseSemver accepts what the semver crate parses: MAJOR.MINOR.PATCH
// without leading zeros, an optional -prerelease whose numeric identifiers
// have no leading zeros, and optional +build metadata.
func parseSemver(s string) (codexSemver, bool) {
	var v codexSemver
	rest := s
	if i := strings.IndexByte(rest, '+'); i >= 0 {
		rest, v.build = rest[:i], rest[i+1:]
		if !validIdentifiers(v.build, false) {
			return codexSemver{}, false
		}
	}
	if i := strings.IndexByte(rest, '-'); i >= 0 {
		rest, v.pre = rest[:i], rest[i+1:]
		if !validIdentifiers(v.pre, true) {
			return codexSemver{}, false
		}
	}
	parts := strings.Split(rest, ".")
	if len(parts) != 3 {
		return codexSemver{}, false
	}
	for i, part := range parts {
		if part == "" || !allDigits(part) || (len(part) > 1 && part[0] == '0') {
			return codexSemver{}, false
		}
		n, err := strconv.ParseUint(part, 10, 64)
		if err != nil {
			return codexSemver{}, false
		}
		v.core[i] = n
	}
	return v, true
}

func validIdentifiers(s string, noLeadingZeros bool) bool {
	for _, id := range strings.Split(s, ".") {
		if id == "" {
			return false
		}
		for _, r := range id {
			if !isASCIIAlnum(r) && r != '-' {
				return false
			}
		}
		if noLeadingZeros && len(id) > 1 && id[0] == '0' && allDigits(id) {
			return false
		}
	}
	return true
}

// comparePrerelease orders a release above any prerelease, then compares
// identifiers: numeric ones numerically and below alphanumeric ones, which
// compare as strings; a longer list wins a tie.
func comparePrerelease(a, b string) int {
	switch {
	case a == b:
		return 0
	case a == "":
		return 1
	case b == "":
		return -1
	}
	return compareIdentifiers(strings.Split(a, "."), strings.Split(b, "."), func(x, y string) int {
		if c := compareInt(len(x), len(y)); c != 0 {
			return c
		}
		return strings.Compare(x, y)
	})
}

// compareBuild orders no metadata below any, then compares identifiers like
// prerelease ones, except that numeric identifiers may have leading zeros:
// 0 < 00 < 1 < 01 < 001 < 2.
func compareBuild(a, b string) int {
	switch {
	case a == b:
		return 0
	case a == "":
		return -1
	case b == "":
		return 1
	}
	return compareIdentifiers(strings.Split(a, "."), strings.Split(b, "."), func(x, y string) int {
		tx, ty := strings.TrimLeft(x, "0"), strings.TrimLeft(y, "0")
		if c := compareInt(len(tx), len(ty)); c != 0 {
			return c
		}
		if c := strings.Compare(tx, ty); c != 0 {
			return c
		}
		return compareInt(len(x), len(y))
	})
}

func compareIdentifiers(a, b []string, numeric func(x, y string) int) int {
	for i, x := range a {
		if i == len(b) {
			return 1
		}
		y := b[i]
		var c int
		switch dx, dy := allDigits(x), allDigits(y); {
		case dx && dy:
			c = numeric(x, y)
		case dx:
			return -1
		case dy:
			return 1
		default:
			c = strings.Compare(x, y)
		}
		if c != 0 {
			return c
		}
	}
	if len(a) == len(b) {
		return 0
	}
	return -1
}

func compareInt(a, b int) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	}
	return 0
}

func allDigits(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func isASCIIAlnum(r rune) bool {
	return r >= '0' && r <= '9' || r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z'
}
