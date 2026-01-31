package skill

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// semver is the regular expression from semver.org without build metadata,
// which a release version doesn't need.
var semver = regexp.MustCompile(`^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)` +
	`(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?$`)

// Version is a strict semantic version.
type Version struct {
	Major, Minor, Patch int
	Pre                 []string
}

// ParseVersion parses a semantic version with an optional "v" prefix.
func ParseVersion(s string) (Version, bool) {
	m := semver.FindStringSubmatch(strings.TrimPrefix(s, "v"))
	if m == nil {
		return Version{}, false
	}
	var (
		v    Version
		errs [3]error
	)
	v.Major, errs[0] = strconv.Atoi(m[1])
	v.Minor, errs[1] = strconv.Atoi(m[2])
	v.Patch, errs[2] = strconv.Atoi(m[3])
	for _, err := range errs {
		if err != nil {
			return Version{}, false
		}
	}
	if m[4] != "" {
		v.Pre = strings.Split(m[4], ".")
	}
	return v, true
}

func (v Version) String() string {
	s := fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
	if len(v.Pre) > 0 {
		s += "-" + strings.Join(v.Pre, ".")
	}
	return s
}

// Compare returns -1, 0 or 1 by semver precedence.
func (v Version) Compare(w Version) int {
	for _, d := range [...]int{v.Major - w.Major, v.Minor - w.Minor, v.Patch - w.Patch} {
		if d != 0 {
			return sign(d)
		}
	}
	if len(v.Pre) == 0 || len(w.Pre) == 0 {
		return sign(len(w.Pre) - len(v.Pre))
	}
	for i := 0; i < len(v.Pre) || i < len(w.Pre); i++ {
		switch {
		case i == len(v.Pre):
			return -1
		case i == len(w.Pre):
			return 1
		}
		p, q := v.Pre[i], w.Pre[i]
		if p == q {
			continue
		}
		np, errP := strconv.Atoi(p)
		nq, errQ := strconv.Atoi(q)
		switch {
		case errP == nil && errQ == nil:
			return sign(np - nq)
		case errP == nil:
			return -1
		case errQ == nil:
			return 1
		}
		return strings.Compare(p, q)
	}
	return 0
}

func sign(n int) int {
	switch {
	case n < 0:
		return -1
	case n > 0:
		return 1
	}
	return 0
}

// DefaultRange is the OctoLab compatibility policy: a skill written for X.Y.Z
// covers binaries from X.Y.Z up to the next breaking release. Before 1.0 a
// minor release is breaking; 0.0.z and prereleases match exactly.
func DefaultRange(v Version) string {
	switch {
	case len(v.Pre) > 0 || (v.Major == 0 && v.Minor == 0):
		return "=" + v.String()
	case v.Major == 0:
		return fmt.Sprintf(">=%s <0.%d.0", v, v.Minor+1)
	}
	return fmt.Sprintf(">=%s <%d.0.0", v, v.Major+1)
}

// Comparator is one condition of a Range.
type Comparator struct {
	Op      string
	Version Version
}

// Range is a list of comparators, all of which must hold.
type Range []Comparator

var comparator = regexp.MustCompile(`^(>=|<=|>|<|=)?(.+)$`)

// ParseRange accepts space-separated comparators: =, >=, >, <= and <.
func ParseRange(s string) (Range, bool) {
	fields := strings.Fields(s)
	if len(fields) == 0 {
		return nil, false
	}
	r := make(Range, 0, len(fields))
	for _, field := range fields {
		m := comparator.FindStringSubmatch(field)
		v, ok := ParseVersion(m[2])
		if !ok || strings.HasPrefix(m[2], "v") {
			return nil, false
		}
		op := m[1]
		if op == "" {
			op = "="
		}
		r = append(r, Comparator{Op: op, Version: v})
	}
	return r, true
}

// Contains follows npm's rule for prereleases: a prerelease matches only
// a range that names a prerelease of the same major.minor.patch.
func (r Range) Contains(v Version) bool {
	if len(v.Pre) > 0 {
		named := false
		for _, c := range r {
			b := c.Version
			named = named || len(b.Pre) > 0 && b.Major == v.Major && b.Minor == v.Minor && b.Patch == v.Patch
		}
		if !named {
			return false
		}
	}
	for _, c := range r {
		d := v.Compare(c.Version)
		ok := map[string]bool{"=": d == 0, ">=": d >= 0, ">": d > 0, "<=": d <= 0, "<": d < 0}[c.Op]
		if !ok {
			return false
		}
	}
	return true
}

// CheckRange enforces the compatibility policy on a declared range: it starts
// at the release itself and ends no later than the next breaking release.
// Narrower is allowed, wider is not.
func CheckRange(v Version, s string) error {
	r, ok := ParseRange(s)
	if !ok {
		return fmt.Errorf("must be a version range")
	}
	exact := len(r) == 1 && r[0].Op == "=" && r[0].Version.Compare(v) == 0
	if exact {
		return nil
	}
	if len(v.Pre) == 0 {
		for _, c := range r {
			if len(c.Version.Pre) > 0 {
				return fmt.Errorf("must not name prereleases for a stable release")
			}
		}
	}
	def := DefaultRange(v)
	limit, _ := ParseRange(def)
	if len(limit) == 1 {
		return fmt.Errorf("must be %s", def)
	}
	if len(r) != 2 || r[0].Op == r[1].Op {
		return fmt.Errorf(`must look like ">=X.Y.Z <A.B.C"`)
	}
	bounds := map[string]Version{r[0].Op: r[0].Version, r[1].Op: r[1].Version}
	lower, hasLower := bounds[">="]
	upper, hasUpper := bounds["<"]
	if !hasLower || !hasUpper {
		return fmt.Errorf(`must look like ">=X.Y.Z <A.B.C"`)
	}
	if lower.Compare(v) != 0 {
		return fmt.Errorf("must start at the release: >=%s", v)
	}
	if upper.Compare(v) <= 0 || upper.Compare(limit[1].Version) > 0 {
		return fmt.Errorf("must end after %s and no later than <%s", v, limit[1].Version)
	}
	return nil
}
