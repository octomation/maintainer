---
code:
id:
databaseId:
number:
url:
title: "workspace: safety: enforce boundaries and report scan obstructions"
labels:
  - "type: bug"
  - "scope: code"
  - "scope: test"
  - "severity: critical"
  - "impact: high"
  - "effort: medium"
milestone: "[[milestone-3]]"
state: OPEN
stateReason:
createdAt: 2026-09-26T09:22:00Z
updatedAt: 2026-09-26T09:22:00Z
lastEditedAt:
closedAt:
---

# workspace: safety: enforce boundaries and report scan obstructions

**Motivation:** the release promise of `fetch` is that it creates and moves checkouts only inside the configured workspace, and the promise of both commands is that every checkout of that workspace is either listed or named as unreachable. A symlink among the parents of a destination, overlapping templates, an unreadable directory under a pin and a repository at an unmanaged level each break one of these today.

**Requirements**

- A clone or move destination with a symlink among its ancestors inside the root is a conflict. The boundary and the occupancy of the destination are checked again immediately before `mkdir`/`rename`, so a path that appears between planning and apply is never written into.
- Collision checks cover actual destinations and existing relevant paths: a destination that is an ancestor or a descendant of another destination or of a tracked path is a conflict for both. `Validate` rejects the demonstrable overlaps between owner and default templates; a general proof over arbitrary templates is not required.
- Recoverable scan errors preserve every reachable result and identify the unreadable path. The incompleteness is observable outside stderr: `status` emits an error row for the obstruction (its path, `status: error`, `error: scan: …`) and exits `1` under the existing rule for error rows; `fetch` emits a report-only `conflict` action with the path and the reason, counted in `summary.conflict`. Pins keep their recursive semantics from the workspace specification §2: the walk stops at discovered checkouts, and no depth bound is introduced.
- A `.git` at a prefix level of the layout is reported as an obstruction row in the same form, and the walk continues past it, so checkouts below stay visible.

**Evidence:** finding 1 of [the gpt-6 review](../reviews/20260926T071223Z.md) (`internal/service/fetch/occupancy.go:29`, `internal/service/git/sync.go:35`) and findings 10 (`internal/service/fetch/planner.go:472`), 15 (`internal/service/workspace/scope.go:89`) and 25 (`internal/service/workspace/scope.go:143`) of [the code review](../reviews/20260926T071417Z.md), all reproduced on `932b9c8`.

**PoC**

The boundary cases are a service-level test, not a shell command, built like the lifecycle tests of `service_test.go`: a fake discoverer, a fake confirmer, the real Git adapter with a disposable local origin (`originWithCommit`, wrapped in `localOriginSync` so the adopter accepts it), an isolated state store, and a temporary root with `root/acme` symlinked to a directory outside the root. Clone: the snapshot `acme/svc` and no checkout. Move: a real checkout under a safe ancestor, `root/old-owner/svc`, tracked in state, and the snapshot transferred to `acme/svc`, so that only the destination's ancestor is the symlink. Expected in both: the plan reports a conflict naming the symlinked destination, apply performs nothing, the source checkout and the outside directory are byte-identical to before, and nothing exists outside the root.

Obstructions in the shell, with the binary built from the branch:

```bash
set -o pipefail; export XDG_STATE_HOME=$(mktemp -d) GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_NOSYSTEM=1
ws=$(mktemp -d) && mkdir -p "$ws/public/o/r" && git -C "$ws/public/o/r" init -q && git -C "$ws/public/o" init -q
maintainer status --config="" --root "$ws" --format=json | jq -c '.[] | {repository, status, error}'
# today: [] and exit 0 — expected: the row for o/r plus an error row naming $ws/public/o as an obstruction, exit 1
```

**Out of scope:** following symlinks during discovery; vendor-directory exclusion lists; a depth bound for pins.

Related: [[issue-445]], [[issue-fetch-preserve-checkout-identity]], [[issue-workspace-resolve-pins-with-context]].
