---
code: MAIN-129
id: I_kwDOE2M9Zc8AAAABUXahqg
databaseId: 5661696426
number: 467
url: https://github.com/octomation/maintainer/issues/467
title: "fetch: adopt: keep adoption of a pinned checkout additive"
issueType: null
labels:
  - "scope: code"
  - "scope: test"
  - "type: bug"
  - "severity: minor"
  - "impact: medium"
assignees:
  - kamilsk
milestone: "[[MAIN-M2]]"
parent: null
issueFields: {}
state: CLOSED
stateReason: COMPLETED
duplicateOf: null
createdAt: 2026-10-01T11:53:51Z
updatedAt: 2026-10-21T10:22:39Z
lastEditedAt: 2026-10-21T05:15:42Z
closedAt: 2026-10-20T12:35:45Z
---

# fetch: adopt: keep adoption of a pinned checkout additive

**Motivation:** the first apply over an unknown tree must only add — state records, and for forks the `upstream` remote — and rearrange nothing; moves and origin updates belong to the next apply over a filled state. That is what makes the two-apply convergence promise of v0.2.0 reviewable: the first plan never hides a mutation of an existing checkout inside an adoption. A pinned checkout is exactly the one the owner relies on most, `~/.dotfiles` among them.

**Requirements**

- `adopt` records the checkout with its observed origin and never changes its remotes.
- A pinned checkout whose origin differs from the canonical URL (the repository was renamed or transferred) is adopted on the first apply and gets a separate, reported `update_remote` on the next one.
- Tests: a pinned checkout of a renamed repository plans `adopt` without a remote change on the first run and `update_remote` on the second; `remote.origin.url` and `remote.origin.pushurl` are byte-identical after the first apply.

**Evidence:** `internal/service/fetch/pinned.go:84` sets `UpdateRemote` on an adoption whenever the observed origin differs from the canonical one, and `internal/service/fetch/applier.go:43-48` then calls `UpdateRemote` inside `KindAdopt` before recording the checkout. The fetch specification says that a move or an origin update after adoption falls to the next apply. Found by the second-round reviews of [[MAIN-128]] (Fable 5.1 and gpt-6-astra, 2026-10-01), confirmed by reading `main`.

**PoC**

A planner test in `internal/service/fetch`: a pin rule for the id of `acme/old`, a snapshot renamed to `acme/new` with the same id, a checkout at the pinned path with origin `acme/old`, no state record. Today the plan is one `adopt` with `UpdateRemote = true`; expected: `adopt` with `UpdateRemote = false`, and after applying it, a second plan with exactly one `update_remote`. A pin rule for the old name `acme/old` without a state record never reaches the adoption: it plans a clone of `acme/new` into the layout, which belongs to [[MAIN-145]].

**Out of scope:** moving a pinned checkout (never done); the dependent fetch after adoption ([[MAIN-136]]).

Related: [[MAIN-145]], [[MAIN-135]], [[MAIN-128]].
