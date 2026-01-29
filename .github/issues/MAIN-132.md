---
code: MAIN-132
id: I_kwDOE2M9Zc8AAAABYd-Omw
databaseId: 5937008283
number: 488
url: https://github.com/octomation/maintainer/issues/488
title: "docs: workspace: publish the fetch and status workflows"
issueType: null
labels:
  - "scope: docs"
  - "type: improvement"
  - "impact: high"
  - "effort: medium"
assignees: []
milestone: "[[MAIN-M3]]"
parent: null
issueFields: {}
state: CLOSED
stateReason: COMPLETED
duplicateOf: null
createdAt: 2026-10-20T06:24:55Z
updatedAt: 2026-10-20T06:25:55Z
lastEditedAt: 2026-10-20T06:25:54Z
closedAt: 2026-10-20T06:24:57Z
---

# docs: workspace: publish the fetch and status workflows

**Motivation:** the site describes `fetch` and `status` as "not released yet; build from source", the home page presents v0.1.0 with three features, and the quickstart does not know the workspace. For v0.2.0, released directly as a stable release, the site, the README and the command help must describe both commands as released, on the scenarios that come from the owner's practice, with examples that reproduce on the built binary, and must claim nothing that the release binary does not do.

**Scenarios**, in this order, from [[MAIN-M3]]: see where local attention is needed; refresh the collection in the weekly routine; bring existing checkouts under management; reconcile names and locations with the vault and the shell helpers; extract focused lists with JSON. Safety is shown inside each scenario, not as a separate story.

**Requirements**

- `docs/content/fetch.md`, `status.md`, `workspace.md`, README: no "not released yet" left; the guides describe v0.2.0 as it ships. A behaviour that a patch after v0.2.0 changes is described as v0.2.0 has it, or named among the note's limitations, never as implemented: owner and profile narrowing and filters ([[MAIN-133]]), usage-error exit codes ([[MAIN-131]]), duplicate and orphan rows ([[MAIN-141]]), the selection under a filter and the summary line ([[MAIN-143]]), pins ([[MAIN-145]]), obstructions and symlinked destinations ([[MAIN-144]]), `state prune` ([[MAIN-137]]), case-only renames ([[MAIN-150]]); the onboarding keeps its second apply, since the same-run refresh ([[MAIN-136]]) is not in the release.
- `--owner` and `--profile` narrow only the discovery of new repositories: tracked checkouts and pins of other owners are still confirmed, fetched, moved and have their origin updated. The help of both flags in `maintainer fetch --help` and `fetch.md` say so; neither promises that a narrowed run works only on the named owners or profiles.
- `[filters]` (`exclude_archived`, `exclude_forks`, `exclude_templates`) gate new clones, not the adoption of checkouts already on disk, pinned ones included. `fetch.md`, the template that `maintainer fetch config init` writes and the onboarding example say so.
- `status.md`, next to the offline guarantee: maintainer performs no remote discovery or fetch during `status`; the inspection disables optional locks, lazy fetch, fsmonitor, external diff and text conversion, and otherwise trusts the local Git configuration as `git status` does, so configured filters may run and have their own side effects (finding 29 of [the code review](../reviews/20260926T071417Z.md), resolved by documentation in v0.2.0); the push-lock marker means at least one remote has `pushurl = no_push`, not that every push destination is blocked.
- `fetch.md`: a 404 by id means "not accessible by id: deleted or access lost", `remote-gone` stays as the compatibility label with its check time; "never deletes" becomes "existing checkouts and local work are retained", since fetch prunes remote-tracking refs and `state prune` forgets records, and `maintainer fetch --help` says the same instead of "It never deletes"; plan mode creates the state directory and lock but persists no reconciliation state; JSON apply output is a plan with aggregate errors, not a per-action journal, and a saved plan is not executable.
- `docs/content/index.mdx`: the description, the release pill and the release band present v0.2.0. The scope note no longer says that maintainer does not change your repositories — `fetch --apply` clones, moves checkouts and updates origins; its new wording is neutral and names the value for the user, help with carrying open source contribution, without a slogan. Next to the contribution flow, a path into fetch and status; `quickstart.md` and the README pin examples name the release.
- `docs/content/changelog/v0.2.0.md` and the changelog index: the note opens with "Fetch & Status: refresh your repository collection and see where local work needs attention", shows the `status → fetch → fetch --apply → status` loop and the transfer example, uses the collection numbers as a dated example, lists the limitations — refs only, one workspace, submodules and LFS not fetched, SSH best-effort, PAT only, and those of the first requirement that an ordinary user meets — and hands off to `git sync`, `git refresh` and the vault index as documentation, not features. No release-candidate note.
- Every example of the guides and the note reproduces on the binary of the release commit; `release.mjs render v0.2.0` passes; the static site builds; `docs/content/changelog/_meta.js` and the navigation know the release.

The places in the tree that these requirements touch, as of the triage of 2026-10-19, are listed in [the issue review](../reviews/20261019T105848Z.md#main-132-publish-the-fetch-and-status-workflows).

**PoC**

```bash
set -o pipefail
grep -rn 'not released yet' README.md docs/content/                 # non-empty today; expected: empty
maintainer fetch --help | grep -n 'never deletes\|limit work'        # non-empty today; expected: empty
node .github/scripts/release.mjs render v0.2.0 --site-url https://maintainer.octolab.org/ --out /tmp/maintainer-v0.2.0.md
TARGET=static SITE_URL=https://maintainer.octolab.org/ ./Taskfile docs build
```

**Out of scope:** documenting features that are not in the release; a docs search index; the code changes of [[MAIN-133]] and the other tasks of a v0.2.x patch.

Related: [[MAIN-138]], [[MAIN-133]], [[MAIN-136]], [[MAIN-143]], [[MAIN-131]], [[MAIN-141]], [[MAIN-145]].
