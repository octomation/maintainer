---
code: MAIN-M3
id: MI_kwDOE2M9Zc4BFBVP
databaseId: 18093391
number: 3
url: https://github.com/octomation/maintainer/milestone/3
title: "Workspace, v0.2"
state: OPEN
openIssueCount: 0
closedIssueCount: 0
dueOn: null
createdAt: 2026-09-26T09:14:44Z
updatedAt: 2026-10-19T07:36:47Z
closedAt: null
---

# Workspace, v0.2

## Workspace: fetch and status

First release of `maintainer fetch` and `maintainer status`: discover and refresh a multi-owner collection of GitHub checkouts, reconcile repository moves, and inspect local work. Plan by default; preserve existing local branches and working files; inspect status without contacting GitHub.

The release ships what `main` already carries, made trustworthy on a real collection: the defects the reviews of 2026-09-26 found in `fetch`, `status` and the workspace are fixed, continuous integration is green again, and the documentation describes both commands as released, on the owner's scenarios. The release path and its acceptance: [[MAIN-138]].

### Value

1. See where local attention is needed — uncommitted work, ahead and diverged branches, unborn repositories, missing upstreams and push locks across the whole collection, offline.
2. Refresh the collection in the weekly routine — review discovery and transfers as a plan, apply only non-destructive actions, then inspect divergence against refreshed refs.
3. Bring existing checkouts under management — adopt without recloning, keep active checkouts outside the tree as pins, see inactive duplicates.
4. Reconcile names and locations with the vault and the shell helpers — stable ids, old and new paths, verified after apply.
5. Extract focused lists with JSON — dirty, locked, orphan, conflict and error lists with dependable exit codes.

### Scope

| Group | Task |
| --- | --- |
| CI and local gates reliable again | [[MAIN-127]] |
| Prereleases stay out of stable channels | [[MAIN-139]] |
| Workspace boundary at mutation time, scan obstructions visible | [[MAIN-144]] |
| Pins and paths only from sufficient context | [[MAIN-145]] |
| State commands honor config selection and preserve uncertain records | [[MAIN-137]] |
| Checkout identity survives renames, transfers and push locks | [[MAIN-135]] |
| Owners, profiles and adoption filters do what they say | [[MAIN-133]] |
| Discovery isolates failures and respects rate limits | [[MAIN-134]] |
| Adoption of a pinned checkout stays additive | [[MAIN-129]] |
| Reconciled checkouts are refreshed in the same apply (scope freeze before the release candidate decides) | [[MAIN-136]] |
| Status rows preserve distinct checkouts and useful errors | [[MAIN-141]] |
| Status summary line and preserved selection | [[MAIN-143]] |
| Status counts: renames detected consistently | [[MAIN-130]] |
| Status counts: re-inspection recomputes, not accumulates | [[MAIN-142]] |
| Usage errors exit with the documented codes | [[MAIN-131]] |
| Documentation and the release note | [[MAIN-132]] |
| The release itself | [[MAIN-138]] |

### Acceptance

Every install channel — archives, `bin/install`, `go install …@latest` and `…@v0.2.0`, Homebrew — delivers a binary with both commands, checked by `maintainer version` and `--help`; on the owner's collection `fetch --apply` preserves the local branch tips, HEAD, index, working files and push URLs of existing checkouts while refs, state and the planned paths and origins are updated, and reconciliation converges: after at most two successful applies against unchanged inputs, the following plan contains no executable reconciliation action, only routine `fetch` and report-only rows; no Med-or-higher finding of the 2026-09-26 reviews on fetch, status or the workspace stays open; CI is green and the release candidate is published as a prerelease that leaves *Latest* and the tap untouched; one checkout's failure leaves the rest usable; existing config, state and pins load unchanged. Formula and Cask are accepted separately before and after the Formula retirement ([[MAIN-125]]).

### Out of scope

Working-tree updates (fast-forward), `upstream` remotes ([[MAIN-128]]), multiple workspaces and linked worktrees, GraphQL discovery, keychain or GitHub App auth, mirror push, submodules and LFS, a fetch TUI, cron mode, a vault transaction journal, the contribution leftovers of milestone 1.

## Issues

- [x] [[MAIN-127]] — ci/cd: verification: restore reliable CI and local gates
- [ ] [[MAIN-129]] — fetch: adopt: keep adoption of a pinned checkout additive
- [x] [[MAIN-130]] — status: inspection: detect renames the same way in every count
- [ ] [[MAIN-131]] — command: exit codes: enforce the documented codes for usage errors
- [ ] [[MAIN-132]] — docs: workspace: publish the fetch and status workflows
- [ ] [[MAIN-133]] — fetch: selection: honor owners, profiles and adoption filters
- [ ] [[MAIN-134]] — fetch: discovery: isolate repository failures and respect rate limits
- [ ] [[MAIN-135]] — fetch: remote: change only the origin URL when updating a remote
- [ ] [[MAIN-136]] — fetch: apply: refresh successfully reconciled checkouts in the same run
- [ ] [[MAIN-137]] — fetch: state: honor config selection and preserve uncertain records
- [ ] [[MAIN-138]] — org: release: prepare v0.2.0
- [ ] [[MAIN-139]] — release: publishing: keep prereleases out of stable channels
- [ ] [[MAIN-141]] — status: inventory: preserve distinct checkouts and useful errors
- [ ] [[MAIN-142]] — status: inspection: recompute counters when a row is inspected again
- [ ] [[MAIN-143]] — status: overview: summarize the inventory and preserve the selection
- [ ] [[MAIN-144]] — workspace: safety: enforce boundaries and report scan obstructions
- [ ] [[MAIN-145]] — workspace: paths: resolve pins only with sufficient context
