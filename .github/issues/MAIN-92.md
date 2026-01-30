---
code: MAIN-92
id: I_kwDOE2M9Zc5biw_k
databaseId: 1535840228
number: 98
url: https://github.com/octomation/maintainer/issues/98
title: "github: setup: add command to manage all available projects and repositories"
issueType: null
labels:
  - "scope: code"
  - "type: feature"
  - "impact: high"
  - "effort: hard"
assignees:
  - kamilsk
milestone: "[[MAIN-M3]]"
parent: null
issueFields: {}
state: CLOSED
stateReason: COMPLETED
duplicateOf: null
createdAt: 2023-01-17T06:23:16Z
updatedAt: 2026-10-20T07:33:36Z
lastEditedAt: 2026-10-19T07:22:52Z
closedAt: 2026-10-19T17:55:47Z
---

# github: setup: add command to manage all available projects and repositories

Manage a local collection of GitHub repositories: discover projects across owners, place clones by comprehensible rules, refresh the data, and see where attention is required. A distinct part of the original idea is configuring the fork relation automatically — `upstream` for the parent.

The refined prototype from the specifications:

```bash
maintainer fetch                 # show the plan
maintainer fetch --apply         # perform the permitted actions
maintainer status --format=plain # the local checkout table
```

The expected result: a new repository is offered for cloning; a rename is recognized by the stable GitHub ID; a clone whose repository disappeared from GitHub stays on disk with a diagnostic. Workspace settings limit the discovery scope, pins select existing checkouts such as `~/.dotfiles`, and local changes are preserved. Status shows the branch, the changes and ahead/behind, problematic copies included.

**Boundaries:** fetch updates refs and performs no merge, reset or push; backing up issues and projects is [[MAIN-31]]. Configuring `upstream` remains an original requirement of this umbrella task, but is not automatically part of the first fetch PoC.

Documented behaviour: [fetch](https://maintainer.octolab.org/fetch/), [status](https://maintainer.octolab.org/status/), [workspace](https://maintainer.octolab.org/workspace/).

<!-- 2026-10-19T17:55Z https://github.com/octomation/maintainer/issues/98#issuecomment-6273696546
Shipped in v0.2.0: `maintainer fetch` discovers repositories across owners, plans clones, adoptions, moves of renamed and transferred repositories by their stable GitHub id and origin updates, and applies them with `--apply`, retaining existing checkouts and local work; `maintainer status` shows uncommitted work, divergence, push locks and inactive duplicates offline. Release notes: https://github.com/octomation/maintainer/releases/tag/v0.2.0. Configuring `upstream` for forks continues in #466.
-->
