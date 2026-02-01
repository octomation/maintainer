package config

import (
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// unknownKey is a key of the decoded document that the Fetch structure, and
// so fetch.cue, does not declare.
type unknownKey struct {
	table   []string // enclosing table; array elements are "name[i]"
	key     string
	isTable bool
	suggest string
}

func (u unknownKey) String() string {
	var b strings.Builder
	switch {
	case u.isTable:
		fmt.Fprintf(&b, "table %s", tableName(append(append([]string(nil), u.table...), quoteKey(u.key))))
	case len(u.table) == 0:
		fmt.Fprintf(&b, "key %s at the top level", quoteKey(u.key))
	default:
		fmt.Fprintf(&b, "key %s in %s", quoteKey(u.key), tableName(u.table))
	}
	switch {
	case u.suggest == "":
	case u.isTable:
		fmt.Fprintf(&b, "; did you mean %s?", tableName(append(append([]string(nil), u.table...), u.suggest)))
	default:
		fmt.Fprintf(&b, "; did you mean %s?", u.suggest)
	}
	return b.String()
}

// unknownKeysError names every unknown key with its table and the file, so a
// typo such as include_owner never silently widens or skips a rule.
func unknownKeysError(path string, keys []unknownKey) error {
	return unknownKeysErrorOf("fetch config", path, keys)
}

func unknownKeysErrorOf(kind, path string, keys []unknownKey) error {
	if len(keys) == 1 {
		return fmt.Errorf("%s %q: unknown %s", kind, path, keys[0])
	}
	lines := make([]string, 0, len(keys))
	for _, k := range keys {
		lines = append(lines, "  "+k.String())
	}
	return fmt.Errorf("%s %q has unknown keys:\n%s", kind, path, strings.Join(lines, "\n"))
}

// unknownKeys walks a generically decoded TOML or YAML document along the
// tags of Fetch and reports, in a stable order, every key it does not declare.
func unknownKeys(doc map[string]any) []unknownKey {
	return unknownKeysOf(doc, reflect.TypeFor[Fetch]())
}

// unknownServiceKeys does the same along the tags of Service.
func unknownServiceKeys(doc map[string]any) []unknownKey {
	return unknownKeysOf(doc, reflect.TypeFor[Service]())
}

func unknownKeysOf(doc map[string]any, typ reflect.Type) []unknownKey {
	var out []unknownKey
	checkKeys(nil, doc, typ, &out)
	sort.SliceStable(out, func(i, j int) bool {
		a, b := strings.Join(out[i].table, "."), strings.Join(out[j].table, ".")
		if a != b {
			return a < b
		}
		return out[i].key < out[j].key
	})
	return out
}

func checkKeys(table []string, value any, typ reflect.Type, out *[]unknownKey) {
	for typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}
	switch typ.Kind() {
	case reflect.Struct:
		fields, ok := asMap(value)
		if !ok {
			return // a type mismatch is reported by the typed decode
		}
		known := structKeys(typ)
		for _, key := range sortedKeys(fields) {
			field, declared := known[key]
			if !declared {
				_, isTable := asMap(fields[key])
				*out = append(*out, unknownKey{
					table: table, key: key, isTable: isTable,
					suggest: closestKey(key, known),
				})
				continue
			}
			checkKeys(append(append([]string(nil), table...), quoteKey(key)), fields[key], field.Type, out)
		}
	case reflect.Map:
		entries, ok := asMap(value)
		if !ok {
			return
		}
		for _, name := range sortedKeys(entries) {
			checkKeys(append(append([]string(nil), table...), quoteKey(name)), entries[name], typ.Elem(), out)
		}
	case reflect.Slice:
		items, ok := value.([]any)
		if !ok || len(table) == 0 {
			return
		}
		last := table[len(table)-1]
		for i, item := range items {
			elem := append(append([]string(nil), table[:len(table)-1]...), fmt.Sprintf("%s[%d]", last, i))
			checkKeys(elem, item, typ.Elem(), out)
		}
	default:
		// Scalars and the reserved `workspaces` (any) have no keys to check.
	}
}

// structKeys maps the declared keys of a config struct to their fields. The
// TOML and YAML tags of every field are the same key.
func structKeys(typ reflect.Type) map[string]reflect.StructField {
	keys := make(map[string]reflect.StructField, typ.NumField())
	for i := range typ.NumField() {
		field := typ.Field(i)
		name, _, _ := strings.Cut(field.Tag.Get("toml"), ",")
		if name == "" || name == "-" {
			continue
		}
		keys[name] = field
	}
	return keys
}

// asMap accepts both TOML (string keys) and yaml.v2 (any keys) tables.
func asMap(value any) (map[string]any, bool) {
	switch m := value.(type) {
	case map[string]any:
		return m, true
	case map[any]any:
		out := make(map[string]any, len(m))
		for k, v := range m {
			out[fmt.Sprint(k)] = v
		}
		return out, true
	default:
		return nil, false
	}
}

func sortedKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

var bareKey = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

func quoteKey(key string) string {
	if bareKey.MatchString(key) {
		return key
	}
	return strconv.Quote(key)
}

func tableName(table []string) string {
	return "[" + strings.Join(table, ".") + "]"
}

// closestKey suggests the declared key a typo was probably meant to be: the
// nearest by edit distance, ignoring case, within a third of its length.
func closestKey(key string, known map[string]reflect.StructField) string {
	best, bestDist := "", -1
	for _, candidate := range sortedFieldKeys(known) {
		dist := editDistance(strings.ToLower(key), candidate)
		if bestDist < 0 || dist < bestDist {
			best, bestDist = candidate, dist
		}
	}
	limit := max(1, len(best)/3)
	if bestDist < 0 || bestDist > limit {
		return ""
	}
	return best
}

func sortedFieldKeys(known map[string]reflect.StructField) []string {
	keys := make([]string, 0, len(known))
	for k := range known {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// editDistance is the optimal string alignment distance between a and b:
// insertions, deletions, substitutions and adjacent transpositions.
func editDistance(a, b string) int {
	ra, rb := []rune(a), []rune(b)
	d := make([][]int, len(ra)+1)
	for i := range d {
		d[i] = make([]int, len(rb)+1)
		d[i][0] = i
	}
	for j := range d[0] {
		d[0][j] = j
	}
	for i := 1; i <= len(ra); i++ {
		for j := 1; j <= len(rb); j++ {
			cost := 1
			if ra[i-1] == rb[j-1] {
				cost = 0
			}
			d[i][j] = min(d[i-1][j]+1, d[i][j-1]+1, d[i-1][j-1]+cost)
			if i > 1 && j > 1 && ra[i-1] == rb[j-2] && ra[i-2] == rb[j-1] {
				d[i][j] = min(d[i][j], d[i-2][j-2]+1)
			}
		}
	}
	return d[len(ra)][len(rb)]
}
