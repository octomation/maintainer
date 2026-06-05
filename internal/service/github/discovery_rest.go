package github

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sort"

	"github.com/google/go-github/v88/github"
	"golang.org/x/oauth2"
)

// perPage is the maximum page size GitHub allows (§9).
const perPage = 100

// OwnerWildcard, an empty include_owners, or "*" selects the authenticated user
// plus every organisation they are a member of (§5.3 extension).
const OwnerWildcard = "*"

// isAllOwners reports whether owners requests the "all" set.
func isAllOwners(owners []string) bool {
	if len(owners) == 0 {
		return true
	}
	for _, o := range owners {
		if o == OwnerWildcard {
			return true
		}
	}
	return false
}

// expandAllOwners returns the authenticated user plus their member orgs,
// sorted for determinism.
func expandAllOwners(login string, members map[string]bool) []string {
	orgs := make([]string, 0, len(members))
	for org := range members {
		orgs = append(orgs, org)
	}
	sort.Strings(orgs)
	return append([]string{login}, orgs...)
}

// ClientFactory builds a go-github client for a profile's PAT. The fetch
// service injects this; contract tests inject a factory over a fake transport.
type ClientFactory func(ctx context.Context, token string) *github.Client

// DefaultClientFactory builds an oauth2-authenticated client whose transport is
// wrapped by the RateGuard (§8.3).
func DefaultClientFactory(ctx context.Context, token string) *github.Client {
	src := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	httpClient := oauth2.NewClient(ctx, src)
	httpClient.Transport = NewRateGuard(httpClient.Transport)
	client, _ := github.NewClient(github.WithHTTPClient(httpClient))
	return client
}

// RESTDiscoverer implements Discoverer and Confirmer over the GitHub REST API.
// It is the only place that depends on go-github (§9).
type RESTDiscoverer struct {
	factory ClientFactory
}

// NewRESTDiscoverer builds a discoverer from a client factory.
func NewRESTDiscoverer(factory ClientFactory) *RESTDiscoverer {
	if factory == nil {
		factory = DefaultClientFactory
	}
	return &RESTDiscoverer{factory: factory}
}

var _ Discoverer = (*RESTDiscoverer)(nil)
var _ Confirmer = (*RESTDiscoverer)(nil)
var _ NameResolver = (*RESTDiscoverer)(nil)

// List enumerates repositories visible to one profile, picking the REST
// endpoint per owner relationship (§5.3), de-duplicating by id within the
// profile (broader visibility wins, §9), and applying the include_owners
// allowlist (§5.3).
func (d *RESTDiscoverer) List(ctx context.Context, profile Profile) (Discovery, error) {
	client := d.factory(ctx, profile.Token)

	me, _, err := client.Users.Get(ctx, "")
	if err != nil {
		return Discovery{}, fmt.Errorf("profile %q: authenticate: %w", profile.Name, err)
	}
	login := me.GetLogin()
	members, err := d.memberships(ctx, client)
	if err != nil {
		return Discovery{}, fmt.Errorf("profile %q: list memberships: %w", profile.Name, err)
	}

	// An empty list or "*" means "the authenticated user + every org they are a
	// member of" (§5.3 extension), driven by the /user/orgs membership set.
	owners := profile.Owners
	if isAllOwners(owners) {
		owners = expandAllOwners(login, members)
	}

	allow := make(map[string]bool, len(owners))
	for _, o := range owners {
		allow[o] = true
	}

	byID := make(map[int64]RepoSnapshot)
	var stats []EndpointStat
	collect := func(endpoint string, repos []*github.Repository) {
		stats = append(stats, EndpointStat{Endpoint: endpoint, Count: len(repos)})
		for _, r := range repos {
			merge(byID, mapRepo(r, profile.Name))
		}
	}

	for _, owner := range owners {
		var (
			endpoint string
			repos    []*github.Repository
			err      error
		)
		switch {
		case owner == login:
			endpoint = "/user/repos"
			repos, err = listSelf(ctx, client)
		case members[owner]:
			endpoint = fmt.Sprintf("/orgs/%s/repos?type=all", owner)
			repos, err = listOrg(ctx, client, owner, "all")
		default:
			endpoint = fmt.Sprintf("/orgs/%s/repos?type=public", owner)
			repos, err = listOrg(ctx, client, owner, "public")
			if isNotFound(err) { // not an org → treat as a plain user
				endpoint = fmt.Sprintf("/users/%s/repos?type=owner", owner)
				repos, err = listUser(ctx, client, owner)
			}
		}
		if err != nil {
			return Discovery{}, fmt.Errorf("profile %q: list %s: %w", profile.Name, owner, err)
		}
		collect(endpoint, repos)
	}

	snapshots := make([]RepoSnapshot, 0, len(byID))
	for _, s := range byID {
		if allow[s.Owner] { // include_owners allowlist after discovery
			snapshots = append(snapshots, s)
		}
	}
	return Discovery{Profile: profile.Name, Endpoints: stats, Snapshots: snapshots}, nil
}

// ConfirmByID re-verifies a single repository by its stable id (§10). The
// caller classifies the returned error (404 → orphan, 401/403 → inaccessible…).
func (d *RESTDiscoverer) ConfirmByID(ctx context.Context, profile Profile, id int64) (RepoSnapshot, error) {
	client := d.factory(ctx, profile.Token)
	repo, _, err := client.Repositories.GetByID(ctx, id)
	if err != nil {
		return RepoSnapshot{}, err
	}
	return mapRepo(repo, profile.Name), nil
}

// ResolveByName follows GitHub's rename redirect to recover the stable id (§4.4).
func (d *RESTDiscoverer) ResolveByName(ctx context.Context, profile Profile, owner, name string) (RepoSnapshot, error) {
	client := d.factory(ctx, profile.Token)
	repo, _, err := client.Repositories.Get(ctx, owner, name)
	if err != nil {
		return RepoSnapshot{}, err
	}
	return mapRepo(repo, profile.Name), nil
}

func (d *RESTDiscoverer) memberships(ctx context.Context, client *github.Client) (map[string]bool, error) {
	out := make(map[string]bool)
	opt := &github.ListOptions{PerPage: perPage}
	for {
		orgs, resp, err := client.Organizations.List(ctx, "", opt)
		if err != nil {
			return nil, err
		}
		for _, o := range orgs {
			out[o.GetLogin()] = true
		}
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	return out, nil
}

func listSelf(ctx context.Context, client *github.Client) ([]*github.Repository, error) {
	var all []*github.Repository
	opt := &github.RepositoryListOptions{
		Affiliation: "owner,collaborator",
		ListOptions: github.ListOptions{PerPage: perPage},
	}
	for {
		repos, resp, err := client.Repositories.List(ctx, "", opt)
		if err != nil {
			return nil, err
		}
		all = append(all, repos...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	return all, nil
}

func listOrg(ctx context.Context, client *github.Client, org, typ string) ([]*github.Repository, error) {
	var all []*github.Repository
	opt := &github.RepositoryListByOrgOptions{Type: typ, ListOptions: github.ListOptions{PerPage: perPage}}
	for {
		repos, resp, err := client.Repositories.ListByOrg(ctx, org, opt)
		if err != nil {
			return nil, err
		}
		all = append(all, repos...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	return all, nil
}

func listUser(ctx context.Context, client *github.Client, user string) ([]*github.Repository, error) {
	var all []*github.Repository
	opt := &github.RepositoryListByUserOptions{Type: "owner", ListOptions: github.ListOptions{PerPage: perPage}}
	for {
		repos, resp, err := client.Repositories.ListByUser(ctx, user, opt)
		if err != nil {
			return nil, err
		}
		all = append(all, repos...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	return all, nil
}

// merge keeps the snapshot with broader visibility for a duplicated id (§5.1).
func merge(byID map[int64]RepoSnapshot, s RepoSnapshot) {
	if existing, ok := byID[s.ID]; ok {
		if s.Visibility.Rank() <= existing.Visibility.Rank() {
			return
		}
	}
	byID[s.ID] = s
}

func mapRepo(r *github.Repository, profile string) RepoSnapshot {
	visibility := Visibility(r.GetVisibility())
	if visibility == "" { // fallback to repo.private when the field is empty (§4.3)
		if r.GetPrivate() {
			visibility = Private
		} else {
			visibility = Public
		}
	}
	return RepoSnapshot{
		ID:            r.GetID(),
		NodeID:        r.GetNodeID(),
		Owner:         r.GetOwner().GetLogin(),
		Name:          r.GetName(),
		Visibility:    visibility,
		DefaultBranch: r.GetDefaultBranch(),
		HTTPSCloneURL: r.GetCloneURL(),
		SSHCloneURL:   r.GetSSHURL(),
		IsFork:        r.GetFork(),
		IsTemplate:    r.GetIsTemplate(),
		IsArchived:    r.GetArchived(),
		SourceProfile: profile,
	}
}

// HTTPStatus extracts the HTTP status code from a go-github error, or 0 when
// the error is not an API error (network/transport). It lets the fetch service
// classify a confirmation outcome without importing go-github (§10).
func HTTPStatus(err error) int {
	var apiErr *github.ErrorResponse
	if errors.As(err, &apiErr) && apiErr.Response != nil {
		return apiErr.Response.StatusCode
	}
	return 0
}

func isNotFound(err error) bool { return HTTPStatus(err) == http.StatusNotFound }
