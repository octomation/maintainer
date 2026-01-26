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

The milestone carries the v0.2 line: v0.2.0 and the patches after it. v0.2.0 ships what `main` already carries, directly as a stable release without a release candidate, made trustworthy on a real collection: the only code fix that the owner requires before the tag — an origin update must not rewrite the rest of the checkout's Git configuration — lands, continuous integration is green again, and the documentation, the help and the note describe both commands as released, on the owner's scenarios, claiming only what the release binary does. The other defects that the reviews of 2026-09-26 and the triage of 2026-10-19 found in `fetch`, `status`, the workspace and the release path stay in this milestone and follow in v0.2.x patches; [the issue review of 2026-10-19](../reviews/20261019T105848Z.md) says, in its column Release, which task goes into v0.2.0. The release path and its acceptance: [[MAIN-138]].

### Value

1. See where local attention is needed — uncommitted work, ahead and diverged branches, unborn repositories, missing upstreams and push locks across the whole collection, offline.
2. Refresh the collection in the weekly routine — review discovery and transfers as a plan, apply only non-destructive actions, then inspect divergence against refreshed refs.
3. Bring existing checkouts under management — adopt without recloning, keep active checkouts outside the tree as pins, see inactive duplicates.
4. Reconcile names and locations with the vault and the shell helpers — stable ids, old and new paths, verified after apply.
5. Extract focused lists with JSON — dirty, locked, orphan, conflict and error lists with dependable exit codes.

### Scope of v0.2.0

| Group | Task |
| --- | --- |
| CI and local gates reliable again | [[MAIN-127]] |
| Status counts: renames detected consistently | [[MAIN-130]] |
| An origin update changes only the origin fetch URL | [[MAIN-135]] |
| The tap is written through the OctoLab Releaser App, in verified commits | [[MAIN-140]] |
| Documentation, help and the release note true for the release binary | [[MAIN-132]] |
| The release itself | [[MAIN-138]] |

Every other task of the list below goes into a v0.2.x patch.

### Acceptance

v0.2.0: every install channel — archives, `bin/install`, `go install …@latest` and `…@v0.2.0`, Homebrew — delivers a binary with both commands, checked by `maintainer version` and `--help`; on the owner's collection `fetch --apply` preserves the local branch tips, HEAD, index, working files and push URLs of existing checkouts while refs, state and the planned paths and origins are updated, and an origin update leaves the rest of the Git configuration as it was; reconciliation converges: after at most two successful applies against unchanged inputs, the following plan contains no executable reconciliation action, only routine `fetch` and report-only rows; every claim of the guides, the help and the note holds on the release binary, and the deferred defects an ordinary user meets are named among the note's limitations; CI is green and v0.2.0 is published as the stable release, *Latest* and the tap; a failed action on one checkout does not stop the others; existing config, state and pins load unchanged. Formula and Cask are accepted separately before and after the Formula retirement ([[MAIN-125]]).

A patch: its tasks meet their criteria, the guides and the help change with the code they describe, and the configuration, state and pins of v0.2.0 load unchanged — no key that v0.2.0 accepted is rejected. Steps 2–3 of [[MAIN-125]] go out only after 2026-11-05 and once the transition is verified.

### Out of scope

A release candidate; working-tree updates (fast-forward), `upstream` remotes ([[MAIN-128]]), multiple workspaces and linked worktrees, GraphQL discovery, keychain or GitHub App auth, mirror push, submodules and LFS, a fetch TUI, cron mode, a vault transaction journal, the contribution leftovers of milestone 1.

## Issues

- [ ] [[MAIN-125]] — release: homebrew: retire the formula in favour of the cask
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
- [ ] [[MAIN-140]] — release: publishing: write the tap through the OctoLab Releaser App
- [ ] [[MAIN-141]] — status: inventory: preserve distinct checkouts and useful errors
- [ ] [[MAIN-142]] — status: inspection: recompute counters when a row is inspected again
- [ ] [[MAIN-143]] — status: overview: summarize the inventory and preserve the selection
- [ ] [[MAIN-144]] — workspace: safety: enforce boundaries and report scan obstructions
- [ ] [[MAIN-145]] — workspace: paths: resolve pins only with sufficient context
- [ ] [[MAIN-146]] — fetch: state: write the state file without following a planted link
- [ ] [[MAIN-147]] — fetch: worktrees: keep linked worktrees usable and read them as Git does
- [ ] [[MAIN-148]] — fetch: clone: do not adopt an interrupted clone as a checkout
- [ ] [[MAIN-149]] — fetch: config: reject unknown keys and keep a CUE schema next to the config
- [ ] [[MAIN-150]] — fetch: identity: preserve checkout identity through repository changes
