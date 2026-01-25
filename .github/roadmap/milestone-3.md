---
code:
id: MI_kwDOE2M9Zc4BFBVP
databaseId: 18093391
number: 3
url: https://github.com/octomation/maintainer/milestone/3
title: "Workspace, v0.2"
state: OPEN
openIssueCount: 0
closedIssueCount: 0
dueOn:
createdAt: 2026-09-26T09:14:44Z
updatedAt: 2026-09-26T09:14:44Z
closedAt:
---

# milestone: Workspace, v0.2

## Workspace: fetch and status

First release of `maintainer fetch` and `maintainer status`: discover and refresh a multi-owner collection of GitHub checkouts, reconcile repository moves, and inspect local work. Plan by default; preserve existing local branches and working files; inspect status without contacting GitHub.

The release ships what `main` already carries, made trustworthy on a real collection: the defects the reviews of 2026-09-26 found in `fetch`, `status` and the workspace are fixed, continuous integration is green again, and the documentation describes both commands as released, on the owner's scenarios. The release path and its acceptance: [[issue-org-prepare-release-v020]].

### Value

1. See where local attention is needed — uncommitted work, ahead and diverged branches, unborn repositories, missing upstreams and push locks across the whole collection, offline.
2. Refresh the collection in the weekly routine — review discovery and transfers as a plan, apply only non-destructive actions, then inspect divergence against refreshed refs.
3. Bring existing checkouts under management — adopt without recloning, keep active checkouts outside the tree as pins, see inactive duplicates.
4. Reconcile names and locations with the vault and the shell helpers — stable ids, old and new paths, verified after apply.
5. Extract focused lists with JSON — dirty, locked, orphan, conflict and error lists with dependable exit codes.

### Scope

| Group | Task |
| --- | --- |
| CI and local gates reliable again | [[issue-465]] |
| Prereleases stay out of stable channels | [[issue-release-keep-prereleases-out-of-stable]] |
| Workspace boundary at mutation time, scan obstructions visible | [[issue-workspace-enforce-boundaries]] |
| Pins and paths only from sufficient context | [[issue-workspace-resolve-pins-with-context]] |
| State commands honor config selection and preserve uncertain records | [[issue-fetch-state-honor-config-selection]] |
| Checkout identity survives renames, transfers and push locks | [[issue-fetch-preserve-checkout-identity]] |
| Owners, profiles and adoption filters do what they say | [[issue-fetch-honor-owners-profiles-filters]] |
| Discovery isolates failures and respects rate limits | [[issue-fetch-isolate-discovery-failures]] |
| Adoption of a pinned checkout stays additive | [[issue-467]] |
| Reconciled checkouts are refreshed in the same apply (scope freeze before the release candidate decides) | [[issue-fetch-refresh-reconciled-checkouts]] |
| Status rows preserve distinct checkouts and useful errors | [[issue-status-preserve-distinct-checkouts]] |
| Status summary line and preserved selection | [[issue-status-summarize-and-preserve-selection]] |
| Status counts: renames detected consistently | [[issue-468]] |
| Status counts: re-inspection recomputes, not accumulates | [[issue-status-reset-inspected-counters]] |
| Usage errors exit with the documented codes | [[issue-command-enforce-exit-codes]] |
| Documentation and the release note | [[issue-docs-publish-fetch-and-status]] |
| The release itself | [[issue-org-prepare-release-v020]] |

### Acceptance

Every install channel — archives, `bin/install`, `go install …@latest` and `…@v0.2.0`, Homebrew — delivers a binary with both commands, checked by `maintainer version` and `--help`; on the owner's collection `fetch --apply` preserves the local branch tips, HEAD, index, working files and push URLs of existing checkouts while refs, state and the planned paths and origins are updated, and reconciliation converges: after at most two successful applies against unchanged inputs, the following plan contains no executable reconciliation action, only routine `fetch` and report-only rows; no Med-or-higher finding of the 2026-09-26 reviews on fetch, status or the workspace stays open; CI is green and the release candidate is published as a prerelease that leaves *Latest* and the tap untouched; one checkout's failure leaves the rest usable; existing config, state and pins load unchanged. Formula and Cask are accepted separately before and after the Formula retirement ([[issue-456]]).

### Out of scope

Working-tree updates (fast-forward), `upstream` remotes ([[issue-466]]), multiple workspaces and linked worktrees, GraphQL discovery, keychain or GitHub App auth, mirror push, submodules and LFS, a fetch TUI, cron mode, a vault transaction journal, the contribution leftovers of milestone 1.

## Issues
