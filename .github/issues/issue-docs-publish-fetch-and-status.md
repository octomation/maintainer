---
code:
id:
databaseId:
number:
url:
title: "docs: workspace: publish the fetch and status workflows"
labels:
  - "type: improvement"
  - "scope: docs"
  - "impact: high"
  - "effort: medium"
milestone: "[[milestone-3]]"
state: OPEN
stateReason:
createdAt: 2026-09-26T09:32:00Z
updatedAt: 2026-09-26T09:32:00Z
lastEditedAt:
closedAt:
---

# docs: workspace: publish the fetch and status workflows

**Motivation:** the site describes `fetch` and `status` as "not released yet; build from source", the home page presents v0.1.0 with three features, and the quickstart does not know the workspace. For v0.2.0 the site must describe both commands as released, on the scenarios that come from the owner's practice, with examples that reproduce on the built binary.

**Scenarios**, in this order, from [[milestone-3]]: see where local attention is needed; refresh the collection in the weekly routine; bring existing checkouts under management; reconcile names and locations with the vault and the shell helpers; extract focused lists with JSON. Safety is shown inside each scenario, not as a separate story.

**Requirements**

- `docs/content/fetch.md`, `status.md`, `workspace.md`, README: no "not released yet" left; the contracts that change in this release are described as implemented — CLI narrowing (`--owner`, `--profile`, `--token` after profile selection, the decision table of [[issue-fetch-honor-owners-profiles-filters]]), filters gating clone and adoption, unresolved pins and pin precedence, obstruction rows, case-only renames, exit codes, the summary line, and the same-run refresh only if [[issue-fetch-refresh-reconciled-checkouts]] ships — the scope freeze before the release candidate decides which wording goes in.
- `status.md`, next to the offline guarantee: maintainer performs no remote discovery or fetch during `status`; the inspection disables optional locks, lazy fetch, fsmonitor, external diff and text conversion, and otherwise trusts the local Git configuration as `git status` does, so configured filters may run and have their own side effects (finding 29 of [the code review](../reviews/20260926T071417Z.md), resolved by documentation in v0.2.0); the push-lock marker means at least one remote has `pushurl = no_push`, not that every push destination is blocked.
- `fetch.md`: a 404 by id means "not accessible by id: deleted or access lost", `remote-gone` stays as the compatibility label with its check time; "never deletes" becomes "existing checkouts and local work are retained", since fetch prunes remote-tracking refs and `state prune` forgets records; plan mode creates the state directory and lock but persists no reconciliation state; JSON apply output is a plan with aggregate errors, not a per-action journal, and a saved plan is not executable.
- `docs/content/index.mdx` and `quickstart.md`: a path into fetch and status next to the contribution flow; the README `v0.1.0` pin example updated to the release.
- `docs/content/changelog/v0.2.0.md` and the changelog index: the note opens with "Fetch & Status: refresh your repository collection and see where local work needs attention", shows the `status → fetch → fetch --apply → status` loop and the transfer example, uses the collection numbers as a dated example, lists the limitations — refs only, one workspace, submodules and LFS not fetched, SSH best-effort, PAT only — and hands off to `git sync`, `git refresh` and the vault index as documentation, not features. The release candidate needs its own note, `docs/content/changelog/v0.2.0-rc1.md`, because `release.mjs check` requires the note of the exact tag; it may be a short pointer to the release note.
- Every example of the guides and the notes reproduces on the binary of the release commit; `release.mjs render` passes for both notes; the static site builds; `docs/content/changelog/_meta.js` and the navigation know the release.

**PoC**

```bash
set -o pipefail
grep -rn 'not released yet' README.md docs/content/                 # non-empty today; expected: empty
node .github/scripts/release.mjs render v0.2.0 --site-url https://maintainer.octolab.org/ --out /tmp/maintainer-v0.2.0.md
node .github/scripts/release.mjs render v0.2.0-rc1 --site-url https://maintainer.octolab.org/ --out /tmp/maintainer-v0.2.0-rc1.md
TARGET=static SITE_URL=https://maintainer.octolab.org/ ./Taskfile docs build
```

**Out of scope:** documenting features that are not in the release; a docs search index.

Related: [[issue-org-prepare-release-v020]], [[issue-fetch-honor-owners-profiles-filters]], [[issue-fetch-refresh-reconciled-checkouts]], [[issue-status-summarize-and-preserve-selection]], [[issue-command-enforce-exit-codes]].
