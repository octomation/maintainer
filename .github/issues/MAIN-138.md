---
code: MAIN-138
id: I_kwDOE2M9Zc8AAAABYd-VHw
databaseId: 5937009951
number: 489
url: https://github.com/octomation/maintainer/issues/489
title: "org: release: prepare v0.2.0"
issueType: null
labels:
  - "scope: docs"
  - "type: improvement"
  - "scope: inventory"
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
createdAt: 2026-10-20T06:25:03Z
updatedAt: 2026-10-21T05:15:59Z
lastEditedAt: 2026-10-21T05:15:59Z
closedAt: 2026-10-20T06:25:07Z
---

# org: release: prepare v0.2.0

Publish v0.2.0, the first release of `maintainer fetch` and `maintainer status`, directly as a stable release, so that a user can install it from every documented channel, refresh a multi-owner collection of checkouts and inspect its state, with existing branches and working files preserved. There is no release candidate: the tag is `v0.2.0`, and the defects that do not block it follow in v0.2.x patches. The requirements and the scope are the tasks of [[MAIN-M3]]; this task is the release path itself, as [the release guide](../../AGENTS.md#releasing) describes it.

**Before the tag**

- The milestone [[MAIN-M3]] holds the tasks of the release and of the patches after it, and [the issue review of 2026-10-19](../reviews/20261019T105848Z.md) says which task goes into v0.2.0.
- Scope: the only fix of behaviour before the tag is an origin update that leaves the rest of the Git configuration intact ([[MAIN-135]]); the guides, the help texts, the config template and the note describe the release binary ([[MAIN-132]]). [[MAIN-136]] does not ship, and the documentation keeps the second apply.
- On the release commit: the checks of the release guide — `make verify`, `make go-integration-test` with a token, `make go-dist-check`, `make release-config-check`, and `make doctor` with no `fail` — then a read-only `status` and `fetch` on the real collection, then `fetch --apply` and the preservation check — local branch tips, HEAD, index, working files and push URLs of existing checkouts unchanged; refs, state and the planned paths and origins updated — and the convergence check: after at most two successful applies against unchanged inputs, the following plan contains no executable reconciliation action — no adopt, relocate, move, update_remote or clone — only routine `fetch` and report-only orphan or noop rows.
- Existing practice survives: the current config and state load without manual migration, the four active pins and their inactive duplicates stay distinguished, and a `git-contrib` invocation keeps its stdout contract.
- The note `docs/content/changelog/v0.2.0.md` exists and renders ([[MAIN-132]]).

**The release**

- The tag is created locally and signed, `make release-check TAG=v0.2.0` passes, then `git push --atomic origin main v0.2.0` after the push lock is lifted.
- v0.2.0 becomes *Latest*, and the tap receives the Formula and the Cask 0.2.0 in verified commits signed as the OctoLab Releaser App ([[MAIN-140]]).

**After publication**

- The four archives respond; `bin/install` installs v0.2.0; `go install go.octolab.org/toolset/maintainer@latest` and `…@v0.2.0` work outside a checkout; every route is checked by `maintainer version` and by `maintainer fetch --help` and `maintainer status --help`, not by a successful download alone.
- Homebrew: before the Formula retirement, the Formula and the Cask 0.2.0 are in the tap and `brew upgrade` delivers them, checked in separate installations; after the retirement of [[MAIN-125]], only the Cask. Steps 2–3 of [[MAIN-125]] belong to the first release after 2026-11-05 and only when the transition is verified; they do not hold this release.
- The deployed site describes the release.
- Then the bookkeeping: [[MAIN-92]] is closed with a resolution that names what shipped and links [[MAIN-128]] for the transferred requirement; closing [[MAIN-64]] for its calendar part is independent housekeeping, not a step of this release.

**Out of scope:** a release candidate, and keeping later prereleases out of the stable channels ([[MAIN-139]]); closing the milestone "GCC, v0.1" and [[MAIN-30]]; any feature outside the plan.

Related: [[MAIN-30]] for the release path this one repeats, [[MAIN-127]].
