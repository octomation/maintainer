// Package fetch holds the reconciliation logic for `maintainer fetch`: the
// pure Planner, the path-template renderer, the disk Adopter, the Applier that
// drives the Git port, and the Reporter. It is the orchestration layer that
// ties the GitHub, Git and state ports together (fetch plan §8.3, §9).
package fetch

import (
	"time"

	"go.octolab.org/toolset/maintainer/internal/service/github"
	"go.octolab.org/toolset/maintainer/internal/state"
)

// Kind is the action a single repository resolves to (§7.1).
type Kind string

// Action kinds. Every kind is non-destructive (§7.2). Conflict is a plan
// outcome (a failed repo, reported), not a side effect.
const (
	KindClone        Kind = "clone"
	KindFetch        Kind = "fetch"
	KindMove         Kind = "move"
	KindRelocate     Kind = "relocate"
	KindUpdateRemote Kind = "update_remote"
	KindAdopt        Kind = "adopt"
	KindOrphan       Kind = "orphan"
	KindNoop         Kind = "noop"
	KindConflict     Kind = "conflict"
)

// Confirmation flags carried by orphan/noop actions (§10 error taxonomy).
const (
	FlagFiltered     = "filtered"     // tracked repo excluded from new-clone rules
	FlagInaccessible = "inaccessible" // 401/403: access lost, not gone
	FlagLegalHold    = "legal-hold"   // 451
)

// Action is one planned operation for one repository. The applier needs both
// the source Snapshot (clone/fetch/move) and the prior Record (move/fetch).
type Action struct {
	Kind   Kind
	ID     int64
	NodeID string
	Owner  string
	Name   string

	Path     string // target path for clone/fetch/adopt; orphan location
	FromPath string // move/relocate source
	ToPath   string // move/relocate target
	FromName string // owner/name before a rename (for the move line)

	RemoteURL    string // canonical, credential-free remote
	Transport    string // resolved transport (ssh|https)
	Profile      string // source profile whose creds materialise the clone
	UpdateRemote bool   // a move that also rewrites remote.origin.url

	Filtered bool   // noop/fetch flagged "filtered" (§7.1)
	Flag     string // orphan/confirmation flag (inaccessible/legal-hold)
	Reason   string // conflict / report-only explanation

	Snapshot *github.RepoSnapshot
	Record   *state.Record
}

// order is the apply priority (§7.4): adopt/relocate first (pure metadata),
// then update_remote, move, clone, fetch. Report-only kinds sort last and are
// never executed.
func (a Action) order() int {
	switch a.Kind {
	case KindAdopt, KindRelocate:
		return 0
	case KindUpdateRemote:
		return 1
	case KindMove:
		return 2
	case KindClone:
		return 3
	case KindFetch:
		return 4
	default: // orphan, noop, conflict — report-only
		return 5
	}
}

// Executable reports whether the applier performs a disk/network side effect
// for this action. Orphan, noop and conflict are report-only (§7.4).
func (a Action) Executable() bool {
	switch a.Kind {
	case KindClone, KindFetch, KindMove, KindRelocate, KindUpdateRemote, KindAdopt:
		return true
	default:
		return false
	}
}

// DiscoverySummary is the per-profile discovery line for the plan (§7.3).
type DiscoverySummary struct {
	Profile   string                `json:"profile"`
	Endpoints []github.EndpointStat `json:"endpoints,omitempty"`
	Count     int                   `json:"count"`
}

// Plan is the full reconciliation plan for one run.
type Plan struct {
	ID          string             `json:"plan_id"`
	GeneratedAt time.Time          `json:"generated_at"`
	Discoveries []DiscoverySummary `json:"discoveries"`
	Root        string             `json:"-"`
	StatePath   string             `json:"-"`
	StateCount  int                `json:"-"`
	Actions     []Action           `json:"actions"`

	// errorCount carries apply failures into the rendered summary.
	errorCount int
}

// Summary counts actions by kind (§7.3 summary line and JSON object).
type Summary struct {
	Clone        int `json:"clone"`
	Fetch        int `json:"fetch"`
	Move         int `json:"move"`
	Relocate     int `json:"relocate"`
	UpdateRemote int `json:"update_remote"`
	Adopt        int `json:"adopt"`
	Orphan       int `json:"orphan"`
	Noop         int `json:"noop"`
	Conflict     int `json:"conflict"`
	Errors       int `json:"errors"`
}

// Summary tallies the plan's actions by kind.
func (p Plan) Summary() Summary {
	var s Summary
	for _, a := range p.Actions {
		switch a.Kind {
		case KindClone:
			s.Clone++
		case KindFetch:
			s.Fetch++
		case KindMove:
			s.Move++
		case KindRelocate:
			s.Relocate++
		case KindUpdateRemote:
			s.UpdateRemote++
		case KindAdopt:
			s.Adopt++
		case KindOrphan:
			s.Orphan++
		case KindNoop:
			s.Noop++
		case KindConflict:
			s.Conflict++
		}
	}
	s.Errors = p.errorCount
	return s
}
