package fetch

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"

	"go.octolab.org/toolset/maintainer/internal/config"
	"go.octolab.org/toolset/maintainer/internal/pkg/exit"
	gitsvc "go.octolab.org/toolset/maintainer/internal/service/git"
	"go.octolab.org/toolset/maintainer/internal/service/github"
	"go.octolab.org/toolset/maintainer/internal/state"
)

// ResolvedProfile is a profile with its token already resolved and its owner
// set already filtered by the --profile/--owner scope (§5).
type ResolvedProfile struct {
	Name   string
	Token  string
	Owners []string
}

// Deps are the wired-up collaborators a Service needs.
type Deps struct {
	Store      *state.Store
	Discoverer github.Discoverer
	Confirmer  github.Confirmer
	Resolver   github.NameResolver
	GitSync    gitsvc.GitSync
	Reporter   *Reporter
	Clock      func() time.Time
	IDGen      func() string
}

// Service is the fetch use-case façade (§8.2 "fetch service").
type Service struct {
	cnf         *config.Fetch
	profiles    []ResolvedProfile
	root        string
	concurrency int
	deps        Deps

	planner *Planner
	adopter *Adopter
	applier *Applier
}

// NewService wires a fetch service from config, resolved profiles and deps.
func NewService(
	cnf *config.Fetch,
	profiles []ResolvedProfile,
	home, cwd string,
	concurrency int,
	deps Deps,
) (*Service, error) {
	if concurrency < 1 {
		concurrency = 1
	}
	if deps.Clock == nil {
		deps.Clock = time.Now
	}
	if deps.IDGen == nil {
		deps.IDGen = defaultIDGen(deps.Clock)
	}
	renderer, err := NewPathRenderer(cnf.Defaults.Root, home, cwd)
	if err != nil {
		return nil, exit.WithUser(err)
	}
	paths := NewPathResolver(cnf, renderer)

	auth := func(profile, transport string) gitsvc.Auth {
		return gitsvc.Auth{Transport: transport, Token: tokenFor(profiles, profile)}
	}
	svc := &Service{
		cnf: cnf, profiles: profiles, root: renderer.Root(),
		concurrency: concurrency, deps: deps,
		planner: NewPlanner(cnf, paths),
		applier: NewApplier(deps.GitSync, auth, deps.Clock),
	}
	if deps.Resolver != nil && deps.GitSync != nil {
		svc.adopter = NewAdopter(deps.GitSync, &profileNameResolver{
			resolver: deps.Resolver,
			profile:  primaryProfile(profiles),
		})
	}
	return svc, nil
}

// Run executes one plan (and optionally applies it). It owns the run lifecycle:
// lock, load, discover, adopt, confirm, plan, render, [apply, save].
func (s *Service) Run(ctx context.Context, apply bool) error {
	release, err := s.deps.Store.Lock()
	if err != nil {
		return exit.WithUser(err)
	}
	defer func() { _ = release() }()

	st, err := s.deps.Store.Load()
	if err != nil {
		return err // transport/state error → exit 1
	}

	snapshots, discoveries, err := s.discover(ctx)
	if err != nil {
		return err
	}

	var clones []DiskClone
	if s.adopter != nil {
		clones, err = s.adopter.Scan(ctx, s.root, snapshots, s.cnf)
		if err != nil {
			return err
		}
	}

	confirmations := s.confirm(ctx, st, snapshots)

	actions, err := s.planner.Plan(PlanInput{
		Snapshots:     snapshots,
		State:         st,
		Clones:        clones,
		Confirmations: confirmations,
	})
	if err != nil {
		return exit.WithUser(err)
	}

	plan := Plan{
		ID:          s.deps.IDGen(),
		GeneratedAt: s.deps.Clock(),
		Discoveries: discoveries,
		Root:        s.root,
		StatePath:   s.deps.Store.Path(),
		StateCount:  len(st.Repos),
		Actions:     actions,
	}

	if !apply {
		// Plan-only must exit without touching the disk (success criterion §1):
		// no state write, no checkout mutation.
		return s.deps.Reporter.Render(plan, false)
	}

	failed := s.apply(ctx, &plan, st)
	s.touchDiscovery(st)
	if err := s.deps.Store.Save(st); err != nil {
		return err
	}
	if rerr := s.deps.Reporter.Render(plan, true); rerr != nil {
		return rerr
	}
	if failed > 0 {
		return exit.WithPartial(fmt.Errorf("%d action(s) failed during apply", failed))
	}
	return nil
}

// discover runs each profile's Discoverer concurrently (bounded) and merges the
// results cross-profile, broader visibility winning (§5.1).
func (s *Service) discover(ctx context.Context) ([]github.RepoSnapshot, []DiscoverySummary, error) {
	results := make([]github.Discovery, len(s.profiles))
	group, gctx := errgroup.WithContext(ctx)
	group.SetLimit(s.concurrency)
	for i, p := range s.profiles {
		group.Go(func() error {
			s.deps.Reporter.Logf("discovering (profile=%s)…", p.Name)
			d, err := s.deps.Discoverer.List(gctx, github.Profile{Name: p.Name, Token: p.Token, Owners: p.Owners})
			if err != nil {
				return err
			}
			results[i] = d
			return nil
		})
	}
	if err := group.Wait(); err != nil {
		return nil, nil, err
	}

	merged := make(map[int64]github.RepoSnapshot)
	var summaries []DiscoverySummary
	for _, d := range results {
		summaries = append(summaries, DiscoverySummary{Profile: d.Profile, Endpoints: d.Endpoints, Count: d.Count()})
		for _, snap := range d.Snapshots {
			mergeSnapshot(merged, snap)
		}
	}
	snapshots := make([]github.RepoSnapshot, 0, len(merged))
	for _, snap := range merged {
		snapshots = append(snapshots, snap)
	}
	sort.Slice(snapshots, func(i, j int) bool { return snapshots[i].ID < snapshots[j].ID })
	return snapshots, summaries, nil
}

// confirm re-verifies every tracked id missing from discovery (§10). Transient
// errors leave the record unchanged and are surfaced on stderr.
func (s *Service) confirm(ctx context.Context, st *state.State, snapshots []github.RepoSnapshot) map[int64]Confirmation {
	if s.deps.Confirmer == nil {
		return nil
	}
	present := make(map[int64]bool, len(snapshots))
	for _, snap := range snapshots {
		present[snap.ID] = true
	}
	out := make(map[int64]Confirmation)
	for i := range st.Repos {
		rec := st.Repos[i]
		if present[rec.ID] || s.cnf.Ignored(rec.ID, rec.OwnerLogin, rec.Name) {
			continue
		}
		profile := github.Profile{Name: rec.SourceProfile, Token: tokenFor(s.profiles, rec.SourceProfile)}
		if profile.Token == "" {
			profile = github.Profile{Name: primaryProfile(s.profiles).Name, Token: primaryProfile(s.profiles).Token}
		}
		snap, err := s.deps.Confirmer.ConfirmByID(ctx, profile, rec.ID)
		status := classifyConfirm(err)
		c := Confirmation{Status: status}
		if status == ConfirmFound {
			c.Snapshot = &snap
		}
		if status == ConfirmTransient {
			s.deps.Reporter.Logf("confirm %s/%s (id=%d): transient error, leaving unchanged: %v", rec.OwnerLogin, rec.Name, rec.ID, err)
		}
		out[rec.ID] = c
	}
	return out
}

// apply executes the action list in the prescribed order (§7.4): adopt/relocate,
// update_remote, move (all sequential), then clone and fetch (bounded parallel).
func (s *Service) apply(ctx context.Context, plan *Plan, st *state.State) int {
	var (
		failures int
		mu       sync.Mutex
	)
	executable := 0
	for _, act := range plan.Actions {
		if act.Executable() {
			executable++
		}
	}
	s.deps.Reporter.Logf("applying %d actions (concurrency=%d)", executable, s.concurrency)

	fail := func(act Action, err error) {
		mu.Lock()
		failures++
		mu.Unlock()
		s.deps.Reporter.Errorf("%s %s/%s: %v", act.Kind, act.Owner, act.Name, err)
	}

	sequential := func(kinds ...Kind) {
		want := kindSet(kinds)
		for _, act := range plan.Actions {
			if !want[act.Kind] {
				continue
			}
			s.deps.Reporter.Infof("%s %s/%s → %s", act.Kind, act.Owner, act.Name, act.Path)
			if err := s.applier.Execute(ctx, act, st); err != nil {
				fail(act, err)
				continue
			}
			_ = s.deps.Store.Save(st) // atomic-ish per-action durability (§6.3)
		}
	}
	parallel := func(kind Kind) {
		group, gctx := errgroup.WithContext(ctx)
		group.SetLimit(s.concurrency)
		for _, act := range plan.Actions {
			if act.Kind != kind {
				continue
			}
			group.Go(func() error {
				s.deps.Reporter.Infof("%s %s/%s → %s", act.Kind, act.Owner, act.Name, act.Path)
				if err := s.applier.Execute(gctx, act, st); err != nil {
					fail(act, err)
				}
				return nil
			})
		}
		_ = group.Wait()
		_ = s.deps.Store.Save(st)
	}

	sequential(KindAdopt, KindRelocate)
	sequential(KindUpdateRemote)
	sequential(KindMove)
	parallel(KindClone)
	parallel(KindFetch)

	plan.replaceSummaryErrors(failures)
	return failures
}

func (s *Service) touchDiscovery(st *state.State) {
	now := s.deps.Clock()
	if st.Profiles == nil {
		st.Profiles = map[string]time.Time{}
	}
	for _, p := range s.profiles {
		st.Profiles[p.Name] = now
	}
}

// replaceSummaryErrors records apply failures so the rendered summary reflects
// them (the Summary is recomputed from actions, so errors are carried here).
func (p *Plan) replaceSummaryErrors(n int) { p.errorCount = n }

func mergeSnapshot(byID map[int64]github.RepoSnapshot, snap github.RepoSnapshot) {
	existing, ok := byID[snap.ID]
	if !ok {
		byID[snap.ID] = snap
		return
	}
	switch {
	case snap.Visibility.Rank() > existing.Visibility.Rank():
		byID[snap.ID] = snap
	case snap.Visibility.Rank() == existing.Visibility.Rank() && snap.SourceProfile < existing.SourceProfile:
		byID[snap.ID] = snap // lexicographic profile tie-break (§5.1)
	}
}

func classifyConfirm(err error) ConfirmStatus {
	if err == nil {
		return ConfirmFound
	}
	switch github.HTTPStatus(err) {
	case 404:
		return ConfirmGone
	case 401, 403:
		return ConfirmInaccessible
	case 451:
		return ConfirmLegalHold
	default:
		return ConfirmTransient
	}
}

func tokenFor(profiles []ResolvedProfile, name string) string {
	for _, p := range profiles {
		if p.Name == name {
			return p.Token
		}
	}
	return ""
}

func primaryProfile(profiles []ResolvedProfile) ResolvedProfile {
	if len(profiles) == 0 {
		return ResolvedProfile{}
	}
	return profiles[0]
}

func kindSet(kinds []Kind) map[Kind]bool {
	m := make(map[Kind]bool, len(kinds))
	for _, k := range kinds {
		m[k] = true
	}
	return m
}

func defaultIDGen(clock func() time.Time) func() string {
	return func() string { return clock().UTC().Format("20060102T150405.000000000Z") }
}

type profileNameResolver struct {
	resolver github.NameResolver
	profile  ResolvedProfile
}

func (n *profileNameResolver) ResolveByName(ctx context.Context, owner, name string) (int64, error) {
	snap, err := n.resolver.ResolveByName(ctx, github.Profile{Name: n.profile.Name, Token: n.profile.Token}, owner, name)
	if err != nil {
		if github.HTTPStatus(err) == 404 {
			return 0, nil
		}
		return 0, err
	}
	return snap.ID, nil
}
