---
code:
id:
databaseId:
number:
url:
title: "org: release: prepare v0.2.0"
labels:
  - "type: improvement"
  - "scope: docs"
  - "scope: inventory"
  - "impact: high"
  - "effort: medium"
milestone: "[[milestone-3]]"
state: OPEN
stateReason:
createdAt: 2026-09-26T09:33:00Z
updatedAt: 2026-09-26T09:33:00Z
lastEditedAt:
closedAt:
---

# org: release: prepare v0.2.0

Publish v0.2.0, the first release of `maintainer fetch` and `maintainer status`, so that a user can install it from every documented channel, refresh a multi-owner collection of checkouts and inspect its state, with existing branches and working files preserved. The requirements and the scope are the tasks of [[milestone-3]]; this task is the release path itself.

**Before the release candidate**

- The milestone [[milestone-3]] holds every task of the pool, created on GitHub from the local files.
- Scope freeze: decide whether [[issue-fetch-refresh-reconciled-checkouts]] ships, and align the documentation and the milestone text with that decision.
- On the candidate commit: `make test-integration` with a token, `make dist-check`, a read-only `status` and `fetch` on the real collection, then `fetch --apply` and the preservation check — local branch tips, HEAD, index, working files and push URLs of existing checkouts unchanged; refs, state and the planned paths and origins updated — and the convergence check: after at most two successful applies against unchanged inputs, the following plan contains no executable reconciliation action — no adopt, relocate, move, update_remote or clone — only routine `fetch` and report-only orphan or noop rows.
- Existing practice survives: the current config and state load without manual migration, the four active pins and their inactive duplicates stay distinguished, and a `git-contrib` invocation keeps its stdout contract.
- The candidate note `docs/content/changelog/v0.2.0-rc1.md` exists and renders; the tag is created locally, `make release-check TAG=v0.2.0-rc1` passes, then `git push --atomic origin main v0.2.0-rc1` after the push lock is lifted.

**The release candidate**

- `v0.2.0-rc1` must be published as a prerelease that leaves the stable channels untouched: *Latest* and the tap contents compared before and after ([[issue-release-keep-prereleases-out-of-stable]]).
- The owner runs the candidate on the real collection for at least one weekly cycle. One candidate is the expected path, not a maximum: a fix found during the trial needs another candidate and the relevant checks again.

**The release**

- The note `docs/content/changelog/v0.2.0.md` exists and renders ([[issue-docs-publish-fetch-and-status]]); the tag is created locally, `make release-check TAG=v0.2.0` passes, then `git push --atomic origin main v0.2.0`.

**After publication**

- The four archives respond; `bin/install` installs v0.2.0; `go install go.octolab.org/toolset/maintainer@latest` and `…@v0.2.0` work outside a checkout; every route is checked by `maintainer version` and by `maintainer fetch --help` and `maintainer status --help`, not by a successful download alone.
- Homebrew: before the Formula retirement, the Formula and the Cask 0.2.0 are in the tap and `brew upgrade` delivers them, checked in separate installations; after the retirement of [[issue-456]], only the Cask. Steps 2–3 of [[issue-456]] belong to the first release after 2026-11-05 and only when the transition is verified; they do not hold this release.
- The deployed site describes the release.
- Then the bookkeeping: [[issue-98]] is closed with a resolution that names what shipped and links [[issue-466]] for the transferred requirement; closing [[issue-64]] for its calendar part is independent housekeeping, not a step of this release.

**Out of scope:** closing the milestone "GCC, v0.1" and [[issue-30]]; any feature outside the plan.

Related: [[issue-30]] for the release path this one repeats, [[issue-465]].
