---
code: MAIN-M2
id: MI_kwDOE2M9Zc4Ae1gM
databaseId: 8083468
number: 2
url: https://github.com/octomation/maintainer/milestone/2
title: "Workspace"
state: OPEN
openIssueCount: 4
closedIssueCount: 34
dueOn: null
createdAt: 2022-06-15T10:13:13Z
updatedAt: 2026-11-21T09:10:49Z
closedAt: null
---

# Workspace

## Workspace: fetch and status

`maintainer fetch` and `maintainer status`: discover and refresh a multi-owner collection of GitHub checkouts, reconcile repository moves, and inspect local work. Plan by default; preserve existing local branches and working files; inspect status without contacting GitHub.

### Value

1. See where local attention is needed — uncommitted work, ahead and diverged branches, unborn repositories, missing upstreams and push locks across the whole collection, offline.
2. Refresh the collection in the weekly routine — review discovery and transfers as a plan, apply only non-destructive actions, then inspect divergence against refreshed refs.
3. Bring existing checkouts under management — adopt without recloning, keep active checkouts outside the tree as pins, see inactive duplicates.
4. Reconcile names and locations with the vault and the shell helpers — stable ids, old and new paths, verified after apply.
5. Extract focused lists with JSON — dirty, locked, orphan, conflict and error lists with dependable exit codes.

## Issues

- [ ] [[MAIN-19]] — transition entity
- [x] [[MAIN-39]] — org: review and classify issues
- [ ] [[MAIN-48]] — ci/cd: tools: test act for local development
- [x] [[MAIN-92]] — github: setup: add command to manage all available projects and repositories
- [ ] [[MAIN-116]] — maintainance: sync with the go-tool template
- [x] [[MAIN-123]] — fetch: prevent destructive clone reconciliation for occupied targets
- [x] [[MAIN-124]] — fetch: treat an empty upstream as a successful no-op
- [x] [[MAIN-127]] — ci/cd: verification: restore reliable CI and local gates
- [ ] [[MAIN-128]] — fetch: remotes: configure upstream for forks and check remote consistency
- [x] [[MAIN-129]] — fetch: adopt: keep adoption of a pinned checkout additive
- [x] [[MAIN-130]] — status: inspection: detect renames the same way in every count
- [x] [[MAIN-131]] — command: exit codes: enforce the documented codes for usage errors
- [x] [[MAIN-132]] — docs: workspace: publish the fetch and status workflows
- [x] [[MAIN-133]] — fetch: selection: honor owners, profiles and adoption filters
- [x] [[MAIN-134]] — fetch: discovery: isolate repository failures and respect rate limits
- [x] [[MAIN-135]] — fetch: remote: change only the origin URL when updating a remote
- [x] [[MAIN-136]] — fetch: apply: refresh successfully reconciled checkouts in the same run
- [x] [[MAIN-137]] — fetch: state: honor config selection and preserve uncertain records
- [x] [[MAIN-138]] — org: release: prepare v0.2.0
- [x] [[MAIN-139]] — release: publishing: keep prereleases out of stable channels
- [x] [[MAIN-140]] — release: publishing: write the tap through the OctoLab Releaser App
- [x] [[MAIN-141]] — status: inventory: preserve distinct checkouts and useful errors
- [x] [[MAIN-142]] — status: inspection: recompute counters when a row is inspected again
- [x] [[MAIN-143]] — status: overview: summarize the inventory and preserve the selection
- [x] [[MAIN-144]] — workspace: safety: enforce boundaries and report scan obstructions
- [x] [[MAIN-145]] — workspace: paths: resolve pins only with sufficient context
- [x] [[MAIN-146]] — fetch: state: write the state file without following a planted link
- [x] [[MAIN-147]] — fetch: worktrees: keep linked worktrees usable and read them as Git does
- [x] [[MAIN-148]] — fetch: clone: do not adopt an interrupted clone as a checkout
- [x] [[MAIN-149]] — fetch: config: reject unknown keys and keep a CUE schema next to the config
- [x] [[MAIN-150]] — fetch: identity: preserve checkout identity through repository changes
- [x] [[MAIN-152]] — release: homebrew: keep the Formula without a deprecation
- [x] [[MAIN-153]] — fetch: config: reject a broken per-repository path template and report it per repository
- [x] [[MAIN-154]] — skill: ship an agent skill for the maintainer CLI and publish it to octolab/skills
