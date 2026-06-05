package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/pelletier/go-toml/v2"
	"github.com/spf13/afero"
	"gopkg.in/yaml.v2"
)

// Transport enumerates the supported Git clone transports (§4.2).
const (
	TransportSSH   = "ssh"
	TransportHTTPS = "https"
)

// Defaults for the fetch config when a field is omitted (§4.2).
const (
	DefaultRoot        = "."
	DefaultPath        = "{{.Visibility}}/{{.Owner}}/{{.Repo}}"
	DefaultCloneURL    = TransportSSH
	DefaultConcurrency = 4
	DefaultTokenEnv    = "GITHUB_TOKEN"
)

// Fetch is the dedicated fetch.{toml,yaml} configuration (§4.2). It layers on
// top of the viper-managed Tool config but owns the owner/repo-level rules.
type Fetch struct {
	Defaults Defaults           `toml:"defaults" yaml:"defaults"`
	Filters  Filters            `toml:"filters" yaml:"filters"`
	Profiles map[string]Profile `toml:"profiles" yaml:"profiles"`
	Owners   []Owner            `toml:"owners" yaml:"owners"`
	Repos    []Repo             `toml:"repos" yaml:"repos"`

	// Source is the path the config was loaded from; empty in single-run mode.
	Source string `toml:"-" yaml:"-"`
}

// Defaults holds the global knobs (§4.2 [defaults]).
type Defaults struct {
	Root        string `toml:"root" yaml:"root"`
	Path        string `toml:"path" yaml:"path"`
	CloneURL    string `toml:"clone_url" yaml:"clone_url"`
	Concurrency int    `toml:"concurrency" yaml:"concurrency"`
	StateFile   string `toml:"state_file" yaml:"state_file"`
}

// Filters gate only new clone/adopt decisions (§4.2).
type Filters struct {
	ExcludeArchived  bool `toml:"exclude_archived" yaml:"exclude_archived"`
	ExcludeForks     bool `toml:"exclude_forks" yaml:"exclude_forks"`
	ExcludeTemplates bool `toml:"exclude_templates" yaml:"exclude_templates"`
}

// Profile is a (token, owners) pair contributing one HTTP client (§5.1).
type Profile struct {
	TokenEnv      string   `toml:"token_env" yaml:"token_env"`
	TokenFile     string   `toml:"token_file" yaml:"token_file"`
	Token         string   `toml:"token" yaml:"token"`
	IncludeOwners []string `toml:"include_owners" yaml:"include_owners"`
	CloneURL      string   `toml:"clone_url" yaml:"clone_url"`
}

// Owner is a per-owner override of the path/clone_url (§4.2 [[owners]]).
type Owner struct {
	Name     string `toml:"name" yaml:"name"`
	Path     string `toml:"path" yaml:"path"`
	CloneURL string `toml:"clone_url" yaml:"clone_url"`
}

// Repo is a per-repo rule, matched by id or owner/name (§4.2 [[repos]]).
type Repo struct {
	Match    RepoMatch `toml:"match" yaml:"match"`
	Path     string    `toml:"path" yaml:"path"`
	CloneURL string    `toml:"clone_url" yaml:"clone_url"`
	Ignore   bool      `toml:"ignore" yaml:"ignore"`
}

// RepoMatch selects a repository by stable id or by owner/name.
type RepoMatch struct {
	ID    int64  `toml:"id" yaml:"id"`
	Owner string `toml:"owner" yaml:"owner"`
	Name  string `toml:"name" yaml:"name"`
}

// NamedProfile pairs a profile with its declared name.
type NamedProfile struct {
	Name string
	Profile
}

// LoadFetch decodes a fetch config from path, dispatching by file extension
// (§4.2). An empty path yields a config with defaults applied (single-run
// mode); the caller supplies owners and a profile separately (§4.1).
func LoadFetch(fs afero.Fs, path string) (*Fetch, error) {
	cnf := new(Fetch)
	if path != "" {
		raw, err := afero.ReadFile(fs, path)
		if err != nil {
			return nil, fmt.Errorf("read fetch config %q: %w", path, err)
		}
		switch ext := strings.ToLower(filepath.Ext(path)); ext {
		case ".toml":
			if err := toml.Unmarshal(raw, cnf); err != nil {
				return nil, fmt.Errorf("parse TOML fetch config %q: %w", path, err)
			}
		case ".yaml", ".yml":
			if err := yaml.Unmarshal(raw, cnf); err != nil {
				return nil, fmt.Errorf("parse YAML fetch config %q: %w", path, err)
			}
		default:
			return nil, fmt.Errorf("unsupported fetch config extension %q (want .toml/.yaml)", ext)
		}
		cnf.Source = path
	}
	cnf.applyDefaults()
	return cnf, nil
}

func (f *Fetch) applyDefaults() {
	if f.Defaults.Root == "" {
		f.Defaults.Root = DefaultRoot
	}
	if f.Defaults.Path == "" {
		f.Defaults.Path = DefaultPath
	}
	if f.Defaults.CloneURL == "" {
		f.Defaults.CloneURL = DefaultCloneURL
	}
	if f.Defaults.Concurrency == 0 {
		f.Defaults.Concurrency = DefaultConcurrency
	}
	if f.Profiles == nil {
		f.Profiles = make(map[string]Profile)
	}
}

// Validate structurally checks the config (used by `fetch config validate`
// and before every run). It enforces the path-containment rule (§4.3) and
// the transport enum (§4.2).
func (f *Fetch) Validate() error {
	if err := validTransport("defaults.clone_url", f.Defaults.CloneURL); err != nil {
		return err
	}
	if f.Defaults.Concurrency < 1 {
		return fmt.Errorf("defaults.concurrency must be >= 1, got %d", f.Defaults.Concurrency)
	}
	// defaults.path and per-owner templates must stay within root (§4.3).
	if err := withinRoot("defaults.path", f.Defaults.Path); err != nil {
		return err
	}
	for i := range f.Owners {
		o := &f.Owners[i]
		if o.Name == "" {
			return fmt.Errorf("owners[%d].name is required", i)
		}
		if o.Path != "" {
			if err := withinRoot(fmt.Sprintf("owners[%q].path", o.Name), o.Path); err != nil {
				return err
			}
		}
		if err := validTransport(fmt.Sprintf("owners[%q].clone_url", o.Name), o.CloneURL); err != nil {
			return err
		}
	}
	for i := range f.Repos {
		r := &f.Repos[i]
		if r.Match.ID == 0 && (r.Match.Owner == "" || r.Match.Name == "") {
			return fmt.Errorf("repos[%d].match requires id or owner+name", i)
		}
		if err := validTransport(fmt.Sprintf("repos[%d].clone_url", i), r.CloneURL); err != nil {
			return err
		}
	}
	for name, p := range f.Profiles {
		if err := validTransport(fmt.Sprintf("profiles[%q].clone_url", name), p.CloneURL); err != nil {
			return err
		}
	}
	return nil
}

func validTransport(field, value string) error {
	switch value {
	case "", TransportSSH, TransportHTTPS:
		return nil
	default:
		return fmt.Errorf("%s must be %q or %q, got %q", field, TransportSSH, TransportHTTPS, value)
	}
}

// withinRoot rejects templates that can escape the root via "..". Absolute and
// "~" paths are rejected here too: they are only valid for per-repo overrides.
func withinRoot(field, tmpl string) error {
	if filepath.IsAbs(tmpl) || strings.HasPrefix(tmpl, "~") {
		return fmt.Errorf("%s must be relative to root (absolute/~ allowed only for per-repo overrides): %q", field, tmpl)
	}
	cleaned := filepath.Clean(tmpl)
	if cleaned == ".." || strings.HasPrefix(cleaned, ".."+string(filepath.Separator)) {
		return fmt.Errorf("%s escapes root via %q", field, tmpl)
	}
	return nil
}

// AllOwnersWildcard selects the authenticated user plus all member orgs.
const AllOwnersWildcard = "*"

// IsAllOwners reports whether include_owners means "all": an empty list or one
// containing "*". The discoverer then expands it via the membership set (§5.3).
func IsAllOwners(owners []string) bool {
	if len(owners) == 0 {
		return true
	}
	for _, o := range owners {
		if o == AllOwnersWildcard {
			return true
		}
	}
	return false
}

// SortedProfiles returns profiles ordered by name, so token defaulting and the
// cross-profile merge tie-break (§5.1) are deterministic.
func (f *Fetch) SortedProfiles() []NamedProfile {
	names := make([]string, 0, len(f.Profiles))
	for name := range f.Profiles {
		names = append(names, name)
	}
	sort.Strings(names)
	out := make([]NamedProfile, 0, len(names))
	for _, name := range names {
		out = append(out, NamedProfile{Name: name, Profile: f.Profiles[name]})
	}
	return out
}

// OwnerOverride returns the per-owner rule for owner, if any.
func (f *Fetch) OwnerOverride(owner string) (Owner, bool) {
	for i := range f.Owners {
		if f.Owners[i].Name == owner {
			return f.Owners[i], true
		}
	}
	return Owner{}, false
}

// RepoOverride returns the per-repo rule matching id or owner/name. An id match
// wins over an owner/name match because id survives renames (§4.4).
func (f *Fetch) RepoOverride(id int64, owner, name string) (Repo, bool) {
	for i := range f.Repos {
		if id != 0 && f.Repos[i].Match.ID == id {
			return f.Repos[i], true
		}
	}
	for i := range f.Repos {
		m := f.Repos[i].Match
		if m.ID == 0 && m.Owner == owner && m.Name == name {
			return f.Repos[i], true
		}
	}
	return Repo{}, false
}

// CloneURLFor resolves the clone transport with precedence
// defaults → profile → owner → repo (§4.2).
func (f *Fetch) CloneURLFor(profile, owner string, id int64, name string) string {
	transport := f.Defaults.CloneURL
	if p, ok := f.Profiles[profile]; ok && p.CloneURL != "" {
		transport = p.CloneURL
	}
	if o, ok := f.OwnerOverride(owner); ok && o.CloneURL != "" {
		transport = o.CloneURL
	}
	if r, ok := f.RepoOverride(id, owner, name); ok && r.CloneURL != "" {
		transport = r.CloneURL
	}
	return transport
}

// Ignored reports whether a repo rule silences the whole pipeline for this id
// or owner/name (ignore = true, §4.2).
func (f *Fetch) Ignored(id int64, owner, name string) bool {
	r, ok := f.RepoOverride(id, owner, name)
	return ok && r.Ignore
}

// ResolveToken resolves a profile's token following the order in §5.2:
// token_file → token_env → inline token. The returned warn is non-empty when
// an inline token is used. fallbackEnv defaults token_env for the first profile.
func (p Profile) ResolveToken(fs afero.Fs, getenv func(string) string, fallbackEnv string) (token, warn string, err error) {
	if p.TokenFile != "" {
		if err := checkPerm(fs, p.TokenFile, 0o600); err != nil {
			return "", "", err
		}
		raw, err := afero.ReadFile(fs, p.TokenFile)
		if err != nil {
			return "", "", fmt.Errorf("read token_file %q: %w", p.TokenFile, err)
		}
		return strings.TrimSpace(string(raw)), "", nil
	}
	env := p.TokenEnv
	if env == "" {
		env = fallbackEnv
	}
	if env != "" {
		if v := getenv(env); v != "" {
			return v, "", nil
		}
	}
	if p.Token != "" {
		return p.Token, "inline token in config is insecure; prefer token_file or token_env", nil
	}
	return "", "", fmt.Errorf("no token resolved (tried token_file, env %q, inline)", env)
}

func checkPerm(fs afero.Fs, path string, max os.FileMode) error {
	info, err := fs.Stat(path)
	if err != nil {
		return fmt.Errorf("stat %q: %w", path, err)
	}
	if perm := info.Mode().Perm(); perm&^max != 0 {
		return fmt.Errorf("%q permissions %#o are more permissive than %#o", path, perm, max)
	}
	return nil
}
