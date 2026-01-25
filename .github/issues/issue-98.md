---
code: null
id: I_kwDOE2M9Zc5biw_k
databaseId: 1535840228
number: 98
url: https://github.com/octomation/maintainer/issues/98
title: "github: setup: add command to manage all available projects and repositories"
labels:
  - "scope: code"
  - "type: feature"
  - "impact: high"
  - "effort: hard"
milestone: null
state: OPEN
stateReason: null
createdAt: 2023-01-17T06:23:16Z
updatedAt: 2026-10-19T07:22:52Z
lastEditedAt: 2026-10-19T07:22:52Z
closedAt: null
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

**Boundaries:** fetch updates refs and performs no merge, reset or push; backing up issues and projects is [[issue-31]]. Configuring `upstream` remains an original requirement of this umbrella task, but is not automatically part of the first fetch PoC.

Documented behaviour: [fetch](https://maintainer.octolab.org/fetch/), [status](https://maintainer.octolab.org/status/), [workspace](https://maintainer.octolab.org/workspace/).
