package fetch

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
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
// set already filtered by the --profile/--owner scope (§5). The first one is
// the primary profile.
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

	// Selection is the explicit --owner and --profile scope of this
	// invocation; the zero value is a full run.
	Selection Selection
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
	names   *profileNameResolver
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
	renderer, err := NewPathRenderer(cnf.WorkspaceConfig().Root, home, cwd)
	if err != nil {
		return nil, exit.WithUser(err)
	}
	paths := NewPathResolver(cnf, renderer)
	if _, err := paths.Scope(); err != nil {
		return nil, exit.WithUser(err)
	}

	auth := func(profile, transport string) gitsvc.Auth {
		return gitsvc.Auth{Transport: transport, Token: tokenFor(profiles, profile)}
	}
	mutator := deps.GitSync
	if mutator != nil {
		mutator = boundedSync{GitSync: mutator, root: renderer.Root()}
	}
	svc := &Service{
		cnf: cnf, profiles: profiles, root: renderer.Root(),
		concurrency: concurrency, deps: deps,
		planner: NewPlanner(cnf, paths),
		applier: NewApplier(mutator, auth, deps.Clock),
	}
	if deps.Resolver != nil && deps.GitSync != nil {
		svc.names = &profileNameResolver{resolver: deps.Resolver, profile: primaryProfile(profiles)}
		svc.adopter = NewAdopter(deps.GitSync, svc.names).Ignoring(cnf)
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

	deselected := s.deselect(st, snapshots)
	confirmations := s.confirm(ctx, st, snapshots, deselected)
	snapshots = append(snapshots, confirmedRenames(st, confirmations)...)
	snapshots, outsidePins, err := s.explicitSnapshots(ctx, st, snapshots, deselected)
	if err != nil {
		return err
	}
	// A pin that cannot be rendered is not scanned; the planner reports it as
	// a conflict for that repository only.
	var extraPaths []string
	rendered := make(map[int64]bool, len(snapshots))
	for _, snap := range snapshots {
		rec, _ := st.ByID(snap.ID)
		rendered[snap.ID] = true
		if path, err := s.planner.paths.PinnedPath(snap, rec); err == nil && path != "" {
			extraPaths = append(extraPaths, path)
		}
	}
	// Persisted explicit pins remain inspectable when confirmation says the
	// remote is gone/inaccessible. Their local path must not be forgotten. A
	// repository with a snapshot has its one current pin above: the record's
	// older metadata would render a stale path.
	for _, rec := range st.Repos {
		if _, outside := deselected[rec.ID]; outside || rendered[rec.ID] || s.cnf.Ignored(rec.ID, rec.OwnerLogin, rec.Name) {
			continue
		}
		if pin, err := s.planner.paths.PinnedPath(github.RepoSnapshot{ID: rec.ID, Owner: rec.OwnerLogin, Name: rec.Name}, &rec); err == nil && pin != "" {
			extraPaths = append(extraPaths, pin)
		}
	}
	var clones []DiskClone
	if s.adopter != nil {
		s.names.narrow(s.selectedNames(st, deselected))
		clones, err = s.adopter.ScanScoped(ctx, s.planner.paths.scope, snapshots, extraPaths...)
		if err != nil {
			return err
		}
		clones = s.names.dropSkipped(clones)
	}
	// A verified 404 is an orphan observation, not a missing-pin identity
	// conflict. Do not infer this from an absent API listing alone.
	for i := range clones {
		c := &clones[i]
		if c.ID != 0 || c.Owner == "" || c.Origins > 1 || c.Error == incompleteClone {
			continue
		}
		for _, rec := range st.Repos {
			confirmation, ok := confirmations[rec.ID]
			if ok && confirmation.Status == ConfirmGone && strings.EqualFold(c.Owner, rec.OwnerLogin) && strings.EqualFold(c.Name, rec.Name) {
				c.ID, c.Error = rec.ID, ""
				break
			}
		}
	}
	// A checkout at a path recorded for an id, with that record's remote,
	// keeps the id (§4.4) before the guard asks whether it may be another
	// record's unsaved move.
	clones = attribute(st, clones)
	s.guardUnsavedMoves(st, clones)
	// Existing workspace checkouts (especially pins) are explicit local scope,
	// even when their owners are absent from remote discovery for new clones.
	known := make(map[int64]bool)
	for _, snap := range snapshots {
		known[snap.ID] = true
	}
	for i := range clones {
		c := &clones[i]
		if c.ID == 0 || known[c.ID] || c.Error != "" || s.cnf.Ignored(c.ID, c.Owner, c.Name) {
			continue
		}
		if _, outside := deselected[c.ID]; outside {
			continue
		}
		rec, tracked := st.ByID(c.ID)
		if tracked && !s.planner.paths.Authorised(*rec, github.RepoSnapshot{}) {
			continue
		}
		if !tracked && !s.deps.Selection.owner(c.Owner) {
			continue // an untracked checkout of an owner outside the selection
		}
		if confirmation, ok := confirmations[c.ID]; ok && confirmation.Status != ConfirmFound {
			continue
		}
		if s.deps.Confirmer == nil {
			c.Error = "workspace checkout metadata is unavailable; leaving unchanged"
			continue
		}
		profile := primaryProfile(s.profiles)
		if rec, ok := st.ByID(c.ID); ok {
			for _, candidate := range s.profiles {
				if candidate.Name == rec.SourceProfile {
					profile = candidate
					break
				}
			}
		}
		snap, err := s.deps.Confirmer.ConfirmByID(ctx, github.Profile{Name: profile.Name, Token: profile.Token}, c.ID)
		if err != nil || snap.ID != c.ID {
			c.Error = "workspace checkout identity could not be confirmed; leaving unchanged"
			continue
		}
		if !s.deps.Selection.owner(snap.Owner) {
			if tracked {
				deselected[c.ID] = transferredOut(snap)
			}
			continue
		}
		snap.SourceProfile = profile.Name
		snapshots = append(snapshots, snap)
		known[c.ID] = true
	}
	occupancy := s.scanOccupancy(snapshots, st)

	knownPaths := map[string]bool{}
	for _, rec := range st.Repos {
		for _, path := range append(append([]string(nil), rec.PreviousPaths...), rec.Path) {
			if path == "" {
				continue
			}
			if _, err := os.Stat(filepath.Join(path, ".git")); !os.IsNotExist(err) {
				knownPaths[path] = true
			}
		}
	}
	actions, err := s.planner.Plan(PlanInput{
		KnownCheckouts: knownPaths,
		Snapshots:      append(snapshots, outsidePins...),
		State:          st,
		Clones:         clones,
		Confirmations:  confirmations,
		Occupancy:      occupancy,
		Deselected:     deselected,
		Aliases:        s.caseAliases(snapshots, st),
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

	failed, unsaved := s.apply(ctx, &plan, st)
	observed := s.cacheRemoteObservations(st, snapshots, confirmations)
	s.touchDiscovery(st)
	saveErr := s.deps.Store.Save(st)
	if saveErr != nil {
		failed++
		s.saveFailed(saveErr, unsaved, observed...)
		plan.replaceSummaryErrors(failed)
	}
	if rerr := s.deps.Reporter.Render(plan, true); rerr != nil {
		return rerr
	}
	if saveErr != nil && plan.executable() == 0 {
		return fmt.Errorf("state not saved: %w", saveErr) // nothing ran → exit 1
	}
	conflicts := plan.Summary().Conflict
	if failed > 0 || conflicts > 0 {
		return exit.WithPartial(fmt.Errorf(
			"apply incomplete: %d error(s), %d conflict(s) unresolved",
			failed, conflicts,
		))
	}
	return nil
}

// saveFailed reports a state save that failed, naming the repositories whose
// applied change it leaves unrecorded, then those whose remote observation it
// loses. Their checkouts stay as applied; the next run recovers a record only
// where the checkout identity is safe.
func (s *Service) saveFailed(err error, unsaved []string, observed ...string) {
	msg := fmt.Sprintf("save state: %v", err)
	if len(unsaved) > 0 {
		msg += "; not recorded: " + strings.Join(unsaved, ", ")
	}
	applied := make(map[string]bool, len(unsaved))
	for _, name := range unsaved {
		applied[name] = true
	}
	var others []string
	for _, name := range observed {
		if !applied[name] {
			others = append(others, name)
		}
	}
	const shown = 5
	switch {
	case len(others) > shown:
		msg += fmt.Sprintf("; remote observations of %s and %d more", strings.Join(others[:shown], ", "), len(others)-shown)
	case len(others) > 0:
		msg += "; remote observations of " + strings.Join(others, ", ")
	}
	s.deps.Reporter.Errorf("%s", msg)
}

// guardUnsavedMoves keeps a checkout from being bound to another repository
// when its origin is the recorded remote of a record whose path is missing:
// it may be that checkout, moved by an apply whose state save and remote
// update failed, and its old name may have been reused since. The same id
// still relocates, unless another such record shares the origin; any other
// id is a conflict naming the record.
func (s *Service) guardUnsavedMoves(st *state.State, clones []DiskClone) {
	missing := make(map[string][]state.Record)
	for _, rec := range st.Repos {
		if rec.Path == "" || s.cnf.Ignored(rec.ID, rec.OwnerLogin, rec.Name) {
			continue
		}
		if _, err := os.Lstat(rec.Path); !os.IsNotExist(err) {
			continue
		}
		for _, name := range recordedNames(rec) {
			if n := len(missing[name]); n == 0 || missing[name][n-1].ID != rec.ID {
				missing[name] = append(missing[name], rec)
			}
		}
	}
	for i := range clones {
		c := &clones[i]
		if c.Error != "" || c.Owner == "" {
			continue
		}
		for _, rec := range missing[strings.ToLower(c.Owner+"/"+c.Name)] {
			if rec.ID == c.ID {
				continue
			}
			c.Error = fmt.Sprintf("origin %s/%s is recorded for id=%d at %s, which is missing; "+
				"the checkout may be that repository, so it is not bound to another one", c.Owner, c.Name, rec.ID, rec.Path)
			break
		}
	}
}

// Only apply persists remote observations. Offline status labels this cache
// explicitly; a plan never writes it and a missing listing is not a 404. It
// returns the observed repositories, sorted.
func (s *Service) cacheRemoteObservations(st *state.State, snapshots []github.RepoSnapshot, confirmations map[int64]Confirmation) []string {
	now := s.deps.Clock().UTC()
	observed := make(map[string]bool)
	for id, c := range confirmations {
		if c.Status == ConfirmTransient || c.Status == ConfirmUnverified {
			continue // nothing was established about the repository
		}
		if rec, ok := st.ByID(id); ok {
			rec.RemoteStatus, rec.RemoteCheckedAt = "", &now
			if c.Status == ConfirmGone {
				rec.RemoteStatus = "gone"
			}
			observed[rec.OwnerLogin+"/"+rec.Name] = true
		}
	}
	for _, snap := range snapshots {
		if rec, ok := st.ByID(snap.ID); ok {
			rec.RemoteStatus, rec.RemoteCheckedAt = "", &now
			observed[rec.OwnerLogin+"/"+rec.Name] = true
		}
	}
	names := make([]string, 0, len(observed))
	for name := range observed {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// explicitSnapshots resolves per-repo pins outside discovery, so an explicit
// local selection is meaningful without adding an entire remote owner. A pin
// is explicit intent: an id-only one is confirmed once even in a narrowed run,
// and one of an owner outside the selection is returned apart, report-only.
func (s *Service) explicitSnapshots(
	ctx context.Context,
	st *state.State,
	snapshots []github.RepoSnapshot,
	deselected map[int64]string,
) ([]github.RepoSnapshot, []github.RepoSnapshot, error) {
	var outside []github.RepoSnapshot
	for _, rule := range s.cnf.Repos {
		if rule.Ignore || rule.Path == "" {
			continue
		}
		matches := func(id int64, owner, name string) bool {
			if rule.Match.ID != 0 {
				return id == rule.Match.ID
			}
			return strings.EqualFold(owner, rule.Match.Owner) && strings.EqualFold(name, rule.Match.Name)
		}
		found := false
		for _, snap := range snapshots {
			if matches(snap.ID, snap.Owner, snap.Name) {
				found = true
				break
			}
		}
		if found {
			continue
		}
		// Known records already went through confirmation; do not bypass a
		// failed confirmation with a second, differently authenticated request.
		for _, rec := range st.Repos {
			if matches(rec.ID, rec.OwnerLogin, rec.Name) {
				found = true
				break
			}
		}
		if found {
			continue
		}
		if rule.Match.ID == 0 && !s.deps.Selection.owner(rule.Match.Owner) {
			continue // the rule itself names an owner outside the selection
		}
		p := primaryProfile(s.profiles)
		profile := github.Profile{Name: p.Name, Token: p.Token}
		var snap github.RepoSnapshot
		var err error
		switch {
		case rule.Match.ID != 0 && s.deps.Confirmer != nil:
			snap, err = s.deps.Confirmer.ConfirmByID(ctx, profile, rule.Match.ID)
			if err == nil && snap.ID != rule.Match.ID {
				err = fmt.Errorf("ID mismatch")
			}
		case rule.Match.ID == 0 && s.deps.Resolver != nil:
			snap, err = s.deps.Resolver.ResolveByName(ctx, profile, rule.Match.Owner, rule.Match.Name)
		default:
			err = fmt.Errorf("identity resolver is unavailable")
		}
		if err != nil || snap.ID == 0 {
			return nil, nil, fmt.Errorf("cannot verify explicit checkout %q: %v", rule.Path, err)
		}
		snap.SourceProfile = p.Name
		if !s.deps.Selection.owner(snap.Owner) {
			deselected[snap.ID] = reasonOutside
			outside = append(outside, snap)
			continue
		}
		snapshots = append(snapshots, snap)
	}
	return snapshots, outside, nil
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

// confirm re-verifies every selected tracked id missing from discovery (§10).
// Transient errors leave the record unchanged and are surfaced on stderr. Only
// the record's own profile can establish that a repository is gone or out of
// reach; another token leaves it unverified. A confirmation that finds the
// repository under an owner outside the selection deselects it.
func (s *Service) confirm(ctx context.Context, st *state.State, snapshots []github.RepoSnapshot, deselected map[int64]string) map[int64]Confirmation {
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
		if _, outside := deselected[rec.ID]; outside {
			continue
		}
		if !s.planner.paths.Authorised(rec, github.RepoSnapshot{}) {
			continue
		}
		used, own := profileNamed(s.profiles, rec.SourceProfile)
		if !own {
			used = primaryProfile(s.profiles)
		}
		snap, err := s.deps.Confirmer.ConfirmByID(ctx, github.Profile{Name: used.Name, Token: used.Token}, rec.ID)
		status := classifyConfirm(err)
		c := Confirmation{Status: status}
		switch {
		case status == ConfirmFound && !s.deps.Selection.owner(snap.Owner):
			deselected[rec.ID] = transferredOut(snap)
			continue
		case status == ConfirmFound:
			c.Snapshot = &snap
		case !own && (status == ConfirmGone || status == ConfirmInaccessible):
			c.Status = ConfirmUnverified
			c.Reason = fmt.Sprintf("not verified: checked with profile %q, not with its own profile %q", used.Name, rec.SourceProfile)
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
// The fetch that refreshes a reconciled checkout runs only when its
// reconciliation succeeded. It returns the failures, failed saves included, and
// the repositories whose applied change is not in the state file yet.
func (s *Service) apply(ctx context.Context, plan *Plan, st *state.State) (int, []string) {
	var (
		failures int
		unsaved  []string
		mu       sync.Mutex
	)
	s.deps.Reporter.Logf("applying %d actions (concurrency=%d)", plan.executable(), s.concurrency)

	unreconciled := make(map[int64]bool)
	fail := func(act Action, err error) {
		errs, reconciled := []error{err}, false
		var moved *movedWithErrors
		if errors.As(err, &moved) {
			// The move stands: each linked worktree it could not bring back
			// is an error of its own, and only a failed remote update keeps
			// the checkout from its refresh.
			errs, reconciled = moved.errors(), moved.update == nil
		}
		mu.Lock()
		failures += len(errs)
		if act.reconciles() && !reconciled {
			unreconciled[act.ID] = true // its refresh does not run
		}
		mu.Unlock()
		for _, err := range errs {
			s.deps.Reporter.Errorf("%s %s/%s: %v", act.Kind, act.Owner, act.Name, err)
		}
	}
	done := func(act Action) {
		mu.Lock()
		// A refresh follows its reconciliation: name the repository once.
		if name := act.Owner + "/" + act.Name; !slices.Contains(unsaved, name) {
			unsaved = append(unsaved, name)
		}
		mu.Unlock()
	}
	// save persists every change applied so far (§6.3 per-action durability).
	// A failed save is an error, never ignored.
	save := func() {
		mu.Lock()
		defer mu.Unlock()
		if err := s.deps.Store.Save(st); err != nil {
			failures++
			s.saveFailed(err, unsaved)
			return
		}
		unsaved = nil
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
				// A move that stands is recorded although a later step of it
				// failed: it is saved, or named as unrecorded, right away.
				var moved *movedWithErrors
				if errors.As(err, &moved) {
					done(act)
					save()
				}
				continue
			}
			done(act)
			save()
		}
	}
	parallel := func(kind Kind) {
		group, gctx := errgroup.WithContext(ctx)
		group.SetLimit(s.concurrency)
		ran := false
		for _, act := range plan.Actions {
			if act.Kind != kind {
				continue
			}
			if act.After != "" && unreconciled[act.ID] {
				s.deps.Reporter.Infof("skip %s %s/%s: %s failed", act.Kind, act.Owner, act.Name, act.After)
				continue
			}
			ran = true
			group.Go(func() error {
				s.deps.Reporter.Infof("%s %s/%s → %s", act.Kind, act.Owner, act.Name, act.Path)
				if err := s.applier.Execute(gctx, act, st); err != nil {
					fail(act, err)
					return nil
				}
				done(act)
				return nil
			})
		}
		_ = group.Wait()
		if ran {
			save()
		}
	}

	sequential(KindAdopt, KindRelocate)
	sequential(KindUpdateRemote)
	sequential(KindMove)
	parallel(KindClone)
	parallel(KindFetch)

	plan.replaceSummaryErrors(failures)
	return failures, unsaved
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

// executable counts the actions apply runs.
func (p *Plan) executable() int {
	n := 0
	for _, act := range p.Actions {
		if act.Executable() {
			n++
		}
	}
	return n
}

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
	p, _ := profileNamed(profiles, name)
	return p.Token
}

func profileNamed(profiles []ResolvedProfile, name string) (ResolvedProfile, bool) {
	for _, p := range profiles {
		if p.Name == name {
			return p, true
		}
	}
	return ResolvedProfile{}, false
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

	// allow, when set, keeps resolution within the selection; skipped
	// remembers the origins it declined to resolve.
	allow   func(owner, name string) bool
	skipped map[string]bool
}

func (n *profileNameResolver) narrow(allow func(owner, name string) bool) {
	n.allow, n.skipped = allow, make(map[string]bool)
}

// dropSkipped removes the checkouts whose origin was outside the selection:
// this invocation neither manages nor reports them.
func (n *profileNameResolver) dropSkipped(clones []DiskClone) []DiskClone {
	if len(n.skipped) == 0 {
		return clones
	}
	kept := clones[:0]
	for _, c := range clones {
		if c.ID == 0 && n.skipped[strings.ToLower(c.Owner+"/"+c.Name)] {
			continue
		}
		kept = append(kept, c)
	}
	return kept
}

func (n *profileNameResolver) ResolveByName(ctx context.Context, owner, name string) (int64, error) {
	if n.allow != nil && !n.allow(owner, name) {
		n.skipped[strings.ToLower(owner+"/"+name)] = true
		return 0, nil
	}
	snap, err := n.resolver.ResolveByName(ctx, github.Profile{Name: n.profile.Name, Token: n.profile.Token}, owner, name)
	if err != nil {
		if github.HTTPStatus(err) == 404 {
			return 0, nil
		}
		return 0, err
	}
	return snap.ID, nil
}
