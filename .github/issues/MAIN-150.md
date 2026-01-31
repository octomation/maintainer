---
code: MAIN-150
id: null
databaseId: null
number: null
url: null
title: "fetch: identity: preserve checkout identity through repository changes"
issueType: Bug
labels:
  - "type: bug"
  - "scope: code"
  - "scope: test"
  - "severity: major"
  - "impact: high"
  - "effort: medium"
assignees:
  - kamilsk
milestone: "[[MAIN-M3]]"
parent: null
issueFields: {}
state: CLOSED
stateReason: COMPLETED
duplicateOf: null
createdAt: null
updatedAt: null
lastEditedAt: null
closedAt: null
---

# fetch: identity: preserve checkout identity through repository changes

## Problem

The central promise of `fetch` is that a rename or a transfer on GitHub becomes a move of the existing checkout, keyed by the stable repository id, and that the checkout keeps its local branches, its push lock and its identity. The real plan of 2026-09-26 contains exactly such a transfer, `octolab/cdn.octolab.org` → `octopot/cdn.octolab.org`. When a renamed repository's old name is taken by a new repository between two runs, the user's checkout is bound to the new repository, the renamed one is cloned afresh, and the next apply fetches unrelated history into the checkout. A rename that changes only letter case, and a move whose remote update fails, leave the record and the disk apart as well.

## PoC

Name reuse, as a planner and service test with the real Git adapter: a state record of id 1 at `public/acme/old`, its checkout carrying a local branch `work`; snapshots `{1: acme/new, 2: acme/old}`, repository 2 with unrelated history.

```text
today:    apply 1 adopts the checkout as id 2 and clones id 1 into public/acme/new, dropping id 1's previous_paths;
          apply 2 fetches repository 2 into the checkout: origin/master follows unrelated history, work is ahead 2, behind 1
expected: a conflict that names both ids, or a move of id 1 to acme/new; never adopt id=2 plus clone id=1
```

The same with a pin rule for `acme/old`, and with the checkout at a path listed in `previous_paths` of id 1.

A case-only rename: a snapshot renamed `octocat` → `OctoCat` with the checkout at `octocat` on a case-insensitive filesystem must yield a move through a temporary name or a remote update, and a second plan must be `fetch` only.

A failed remote update after a move: a fake Git port whose `UpdateRemote` fails after a successful `Move`; today the record stays at the old path, and the next run plans a `relocate`.

## Expected behavior

- A checkout at a path that a state record attributes to a different id — the record's path or one of its `previous_paths` — is a conflict that names both ids, never an adoption, on the ordinary and on the pinned adoption path alike. The recorded id may win instead when the stored `remote_url` equals the checkout's origin.
- A rename that changes only letter case is not a conflict when `os.SameFile` confirms the destination is the checkout itself and the destination is otherwise safe: the checkout is renamed through a temporary name, or only the remote is updated. After a rename the record carries the new spelling; when only the remote was updated, the record keeps the on-disk spelling and the planner treats paths that are equal ignoring case and confirmed by `os.SameFile` as settled, so the next plan proposes no move. A different checkout at the destination, a failed identity verification or an unsafe destination stays a conflict.
- A move commits its new path to state as soon as the rename succeeds, before the remote update; a failed remote update therefore leaves a record at the new path, and the next run retries `update_remote` without proposing the move again.
- After a successful fetch the record carries the current snapshot's owner, name, visibility, default branch and flags, with `last_seen` updated.

## Acceptance criteria

- [ ] The name-reuse test yields a conflict naming ids 1 and 2, or a move of id 1, and never `adopt id=2` plus `clone id=1`; the same for the pin rule and for a path in `previous_paths`. The checkout's branches, HEAD, index and files are byte-identical afterwards.
- [ ] A case-only rename on a case-insensitive filesystem converges in one apply, and the second plan is `fetch` only; on a case-sensitive filesystem an occupied destination stays a conflict.
- [ ] A remote-only convergence: after a remote update without a rename, the next plan proposes nothing for that checkout but `fetch`.
- [ ] After a successful rename and a failed remote update, the record is at the new path, and the rerun plans only `update_remote`.
- [ ] After an ordinary fetch, a record whose visibility, default branch or flags changed on GitHub carries the new values and a new `last_seen`.

## Out of scope

The origin update itself, which must keep the rest of the Git configuration ([[MAIN-135]]); configuring `upstream` remotes ([[MAIN-128]]); moves across filesystems; submodules.

## Sources

- [[20261019T105848Z#MAIN-135. Checkout identity]] — the name-reuse reproduction.
- [[20260926T071417Z#9. A renamed repository whose old name is reused is re-cloned and its checkout adopted as the new one]] and [[20260926T071417Z#19. A rename that changes only letter case is a permanent conflict on macOS]].
- [[20260710T095546Z#4. Successful fetches leave last-observed repository metadata stale]].
- Split off [[MAIN-135]], which keeps the origin update for v0.2.0.

<!-- 2026-10-01T12:25Z
Preparation for implementation, from the spec reviews of #466 (2026-10-01): reconfirmed on `main` — `applier.go:79-97` runs `Move`, then `UpdateRemote` (returning before any commit on error), then one `commit`, so a failed remote update leaves state at the old path and recovery goes through `relocate`. Separately, `planner.go:288` adopts a workspace checkout with no path cross-check: a clone of a fork's parent at the fork's path is adopted as the parent and moved on the next run. #466 now specifies that a path belonging to another known id is a conflict, not an adoption; this identity work is the natural place for the check.
-->
