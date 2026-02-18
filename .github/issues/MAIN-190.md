---
code: MAIN-190
id: I_kwDOE2M9Zc8AAAABi075PA
databaseId: 6632175932
number: 574
url: https://github.com/octomation/maintainer/issues/574
title: "fetch: root: scan the folder of a working directory reached through a symlink"
issueType: Bug
labels:
  - "effort: easy"
  - "impact: medium"
  - "scope: code"
  - "scope: test"
  - "severity: major"
  - "type: bug"
assignees:
  - kamilsk
milestone: null
parent: null
issueFields: {}
state: OPEN
stateReason: null
duplicateOf: null
createdAt: 2026-11-26T17:42:13Z
updatedAt: 2026-11-26T17:42:14Z
lastEditedAt: null
closedAt: null
---

# fetch: root: scan the folder of a working directory reached through a symlink

## Problem

`fetch` takes the working directory as the path the shell reports. When the
user reached the folder through a symlink and the workspace root is the
working directory (`root = "."` or the single-run mode), the scan refuses the
root as a symlink. `status` follows such a folder since [[MAIN-165]]; `fetch`
does not.

## PoC

```bash
mkdir real && ln -s real link && cd link
maintainer fetch --config= --owner <owner>
# expected today: workspace path …/link is not a directory (symlink roots are not followed)
# expected after: the plan of real, as when run in real
```

## Expected behavior

- `fetch` treats a working directory reached through a symlink as the folder
  the symlink names, as `status` does.
- A root that the user names explicitly as a symlink is still refused.

## Acceptance criteria

- [ ] A test with the implicit root through a symlink plans as in the real
      folder.
- [ ] A test with an explicit symlinked root still gets the refusal.
- [ ] `docs/content/fetch.md` states the rule as `docs/content/status.md` does.

## Out of scope

Symlinks inside the workspace: checkouts reached through links stay as they
are.

Related: [[MAIN-165]].
