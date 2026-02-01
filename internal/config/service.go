package config

import (
	"fmt"
	"math"
	"net"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/pelletier/go-toml/v2"
	"github.com/spf13/afero"
)

// Defaults of the service config when a key is omitted.
const (
	DefaultScoreWeight    = 1.0
	DefaultAgeHorizonDays = 365
	DefaultDueHorizonDays = 30
	DefaultRepoWeight     = 1.0
)

// Service is the service.toml configuration of `maintainer service run`. The
// loader is strict like Fetch: service.cue is the contract and every key
// outside it is an error. Weights are pointers so that an omitted key takes
// its default while an explicit 0 stays 0; Weights and RepoWeight resolve them.
type Service struct {
	// Listen replaces the default loopback addresses; --listen replaces it.
	Listen []string      `toml:"listen" yaml:"listen"`
	Fetch  FetchRef      `toml:"fetch" yaml:"fetch"`
	Score  Score         `toml:"score" yaml:"score"`
	Repos  []ServiceRepo `toml:"repos" yaml:"repos"`

	// Source is the path the config was loaded from; empty for the defaults.
	Source string `toml:"-" yaml:"-"`
}

// FetchRef names the fetch config the service reads its workspace from.
type FetchRef struct {
	// Config is the path of the fetch config, relative to the home directory;
	// empty means the usual lookup with the home directory as working directory.
	Config string `toml:"config" yaml:"config"`
}

// Score holds the weights of the score formula and its horizons.
type Score struct {
	Impact         *float64 `toml:"impact" yaml:"impact"`
	Effort         *float64 `toml:"effort" yaml:"effort"`
	Type           *float64 `toml:"type" yaml:"type"`
	Age            *float64 `toml:"age" yaml:"age"`
	Due            *float64 `toml:"due" yaml:"due"`
	AgeHorizonDays *int     `toml:"age_horizon_days" yaml:"age_horizon_days"`
	DueHorizonDays *int     `toml:"due_horizon_days" yaml:"due_horizon_days"`
}

// ServiceRepo is a per-repository rule: a weight, matched by id or owner/name.
type ServiceRepo struct {
	Match  RepoMatch `toml:"match" yaml:"match"`
	Weight *float64  `toml:"weight" yaml:"weight"`
}

// Weights are the resolved values of Score.
type Weights struct {
	Impact, Effort, Type, Age, Due float64
	AgeHorizonDays, DueHorizonDays int
}

// ServiceConfigLookup locates service.toml: $XDG_CONFIG_HOME/maintainer/
// service.toml, else ~/.config/maintainer/service.toml.
type ServiceConfigLookup struct {
	Explicit    string
	ExplicitSet bool
	Getenv      func(string) string
	Home        string
}

// Resolve returns the config path and whether it must exist: an explicit path
// must, a discovered one is used only when it exists, and no path at all
// means the defaults.
func (l ServiceConfigLookup) Resolve(fs afero.Fs) (path string, required bool) {
	if l.ExplicitSet {
		return l.Explicit, true
	}
	if candidate := l.Default(); candidate != "" && exists(fs, candidate) {
		return candidate, false
	}
	return "", false
}

// Default is the path service.toml has when no --config names one, whether
// or not a file is there; empty when neither XDG_CONFIG_HOME nor the home
// directory is known.
func (l ServiceConfigLookup) Default() string {
	getenv := l.Getenv
	if getenv == nil {
		getenv = func(string) string { return "" }
	}
	base := getenv("XDG_CONFIG_HOME")
	if base == "" && l.Home != "" {
		base = filepath.Join(l.Home, ".config")
	}
	if base == "" {
		return ""
	}
	return filepath.Join(base, "maintainer", "service.toml")
}

// LoadService decodes service.toml from path and applies the defaults. An
// empty path yields the defaults. A missing file is an error: the caller
// decides whether a discovered path may be missing (ServiceConfigLookup).
func LoadService(fs afero.Fs, path string) (*Service, error) {
	cnf := new(Service)
	if path != "" {
		if ext := strings.ToLower(filepath.Ext(path)); ext != ".toml" {
			return nil, fmt.Errorf("unsupported service config extension %q (want .toml)", ext)
		}
		raw, err := afero.ReadFile(fs, path)
		if err != nil {
			return nil, fmt.Errorf("read service config %q: %w", path, err)
		}
		if err := toml.Unmarshal(raw, cnf); err != nil {
			return nil, fmt.Errorf("parse service config %q: %w", path, err)
		}
		var tables map[string]any
		if err := toml.Unmarshal(raw, &tables); err != nil {
			return nil, err
		}
		if unknown := unknownServiceKeys(tables); len(unknown) > 0 {
			return nil, unknownKeysErrorOf("service config", path, unknown)
		}
		cnf.Source = path
	}
	cnf.applyDefaults()
	if err := cnf.Validate(); err != nil {
		if path != "" {
			return nil, fmt.Errorf("service config %q: %w", path, err)
		}
		return nil, err
	}
	return cnf, nil
}

func (s *Service) applyDefaults() {
	weight := func(p **float64) {
		if *p == nil {
			v := DefaultScoreWeight
			*p = &v
		}
	}
	weight(&s.Score.Impact)
	weight(&s.Score.Effort)
	weight(&s.Score.Type)
	weight(&s.Score.Age)
	weight(&s.Score.Due)
	if s.Score.AgeHorizonDays == nil {
		v := DefaultAgeHorizonDays
		s.Score.AgeHorizonDays = &v
	}
	if s.Score.DueHorizonDays == nil {
		v := DefaultDueHorizonDays
		s.Score.DueHorizonDays = &v
	}
	for i := range s.Repos {
		if s.Repos[i].Weight == nil {
			v := DefaultRepoWeight
			s.Repos[i].Weight = &v
		}
	}
}

// Validate checks the values: loopback addresses, finite non-negative
// weights, positive horizons, and a match that names an id or both owner and
// name.
func (s *Service) Validate() error {
	if s.Listen != nil && len(s.Listen) == 0 {
		return fmt.Errorf("listen needs at least one address; omit it for the defaults")
	}
	for i, address := range s.Listen {
		if err := CheckLoopback(address); err != nil {
			return fmt.Errorf("listen[%d] %w", i, err)
		}
	}
	for _, w := range []struct {
		name  string
		value *float64
	}{{"impact", s.Score.Impact}, {"effort", s.Score.Effort}, {"type", s.Score.Type}, {"age", s.Score.Age}, {"due", s.Score.Due}} {
		if w.value == nil {
			continue
		}
		if math.IsNaN(*w.value) || math.IsInf(*w.value, 0) || *w.value < 0 {
			return fmt.Errorf("score.%s must be a finite non-negative number, got %v", w.name, *w.value)
		}
	}
	if s.Score.AgeHorizonDays != nil && *s.Score.AgeHorizonDays <= 0 {
		return fmt.Errorf("score.age_horizon_days must be positive, got %d", *s.Score.AgeHorizonDays)
	}
	if s.Score.DueHorizonDays != nil && *s.Score.DueHorizonDays <= 0 {
		return fmt.Errorf("score.due_horizon_days must be positive, got %d", *s.Score.DueHorizonDays)
	}
	for i, r := range s.Repos {
		if r.Match.ID == 0 && (r.Match.Owner == "" || r.Match.Name == "") {
			return fmt.Errorf("repos[%d].match requires id or owner+name", i)
		}
		if r.Weight != nil && (math.IsNaN(*r.Weight) || math.IsInf(*r.Weight, 0) || *r.Weight < 0) {
			return fmt.Errorf("repos[%d].weight must be a finite non-negative number, got %v", i, *r.Weight)
		}
	}
	return nil
}

// CheckLoopback accepts a literal loopback IP address with a port number:
// an empty port or a service name would bind another port than the one
// written, and 0 picks a free one. The error reads after the setting's name.
func CheckLoopback(address string) error {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return fmt.Errorf("must be an address with a port: %w", err)
	}
	if ip := net.ParseIP(host); ip == nil || !ip.IsLoopback() {
		return fmt.Errorf("must use a literal loopback IP address, not %q", host)
	}
	if _, err := strconv.ParseUint(port, 10, 16); err != nil {
		return fmt.Errorf("must use a port number from 0 to 65535, not %q", port)
	}
	return nil
}

// Weights returns the resolved score weights and horizons.
func (s *Service) Weights() Weights {
	s.applyDefaults()
	return Weights{
		Impact: *s.Score.Impact, Effort: *s.Score.Effort, Type: *s.Score.Type, Age: *s.Score.Age, Due: *s.Score.Due,
		AgeHorizonDays: *s.Score.AgeHorizonDays, DueHorizonDays: *s.Score.DueHorizonDays,
	}
}

// RepoWeight returns the weight of a repository: the first rule matching by
// id wins, else the first matching by owner and name, case-insensitively, else 1.
func (s *Service) RepoWeight(id int64, owner, name string) float64 {
	for _, r := range s.Repos {
		if id != 0 && r.Match.ID == id {
			return *r.Weight
		}
	}
	for _, r := range s.Repos {
		if r.Match.ID == 0 && strings.EqualFold(r.Match.Owner, owner) && strings.EqualFold(r.Match.Name, name) {
			return *r.Weight
		}
	}
	return DefaultRepoWeight
}

// FetchConfigPath resolves the fetch config the service reads: [fetch] config
// against home when relative, else the usual lookup from home. The second
// result reports whether the path was named explicitly and so must exist.
func (s *Service) FetchConfigPath(fs afero.Fs, getenv func(string) string, home string) (path string, explicit bool) {
	if s.Fetch.Config != "" {
		return ResolveAgainst(home, s.Fetch.Config), true
	}
	path, _ = FetchConfigLookup{Getenv: getenv, WorkDir: home, Home: home}.Resolve(fs)
	return path, false
}

// ResolveAgainst makes a relative path absolute against base; "~/" means home,
// which base is for the service.
func ResolveAgainst(base, path string) string {
	if path == "~" || strings.HasPrefix(path, "~/") {
		return filepath.Join(base, strings.TrimPrefix(path, "~"))
	}
	if filepath.IsAbs(path) {
		return filepath.Clean(path)
	}
	return filepath.Join(base, path)
}
