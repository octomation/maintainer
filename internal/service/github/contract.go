package github

import "context"

//go:generate mockgen -source $GOFILE -destination mocks_test.go -package ${GOPACKAGE}_test

// Visibility classifies a repository's exposure on GitHub.
type Visibility string

// Known repository visibility values (fetch plan §4.3, §6.2).
const (
	Public   Visibility = "public"
	Private  Visibility = "private"
	Internal Visibility = "internal"
)

// Rank orders visibility by breadth: a broader visibility supersedes a
// narrower one when the same repository is seen from two profiles (§5.1).
func (v Visibility) Rank() int {
	switch v {
	case Internal:
		return 3
	case Private:
		return 2
	case Public:
		return 1
	default:
		return 0
	}
}

// RepoSnapshot is the API-side view of a single repository, produced by the
// Discoverer port. The only stable key is ID; every other field is a
// last-observed value (§4.4 identity invariant).
type RepoSnapshot struct {
	ID            int64
	NodeID        string
	Owner         string // repo.owner.login
	Name          string
	Visibility    Visibility
	DefaultBranch string
	HTTPSCloneURL string // https://github.com/owner/repo.git
	SSHCloneURL   string // git@github.com:owner/repo.git
	IsFork        bool
	IsTemplate    bool
	IsArchived    bool

	// SourceProfile is the profile whose credentials surfaced this snapshot.
	// It is filled by the discovery merge, not by the REST decode (§5.1).
	SourceProfile string
}

// FullName returns the canonical owner/name identifier.
func (s RepoSnapshot) FullName() string { return s.Owner + "/" + s.Name }

// Profile is the (token, owners, transport) tuple a Discoverer needs to
// enumerate one slice of GitHub. It mirrors a config profile but stays free
// of the config package so the GitHub port owns no non-GitHub contracts.
type Profile struct {
	Name     string
	Token    string
	Owners   []string // include_owners allowlist (§5.3)
	Excluded map[int64]bool
}

// EndpointStat records how many repositories one REST endpoint contributed,
// for the discovery summary line (§7.3).
type EndpointStat struct {
	Endpoint string
	Count    int
}

// Discovery is the per-profile outcome of Discoverer.List.
type Discovery struct {
	Profile   string
	Endpoints []EndpointStat
	Snapshots []RepoSnapshot
}

// Count returns the number of distinct repositories discovered.
func (d Discovery) Count() int { return len(d.Snapshots) }

// Discoverer enumerates repositories visible to one profile. The REST
// implementation lives in discovery_rest.go; a GraphQL one is deferred to
// milestone 8 (§12). Both produce the same RepoSnapshot DTO.
type Discoverer interface {
	List(ctx context.Context, profile Profile) (Discovery, error)
}

// Confirmer re-verifies a single repository by its stable id, used to confirm
// a disappearance before it is reported as an orphan (§10).
type Confirmer interface {
	ConfirmByID(ctx context.Context, profile Profile, id int64) (RepoSnapshot, error)
}

// NameResolver resolves an owner/name to a snapshot following GitHub's rename
// redirect (GET /repos/{owner}/{name}), recovering the stable id (§4.4).
type NameResolver interface {
	ResolveByName(ctx context.Context, profile Profile, owner, name string) (RepoSnapshot, error)
}
