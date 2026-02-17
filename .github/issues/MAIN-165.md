---
code: MAIN-165
id: I_kwDOE2M9Zc8AAAABihzadA
databaseId: 6612114036
number: 551
url: https://github.com/octomation/maintainer/issues/551
title: "status: root: scan the folder of a symlinked working directory"
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
state: CLOSED
stateReason: COMPLETED
duplicateOf: null
createdAt: 2026-11-25T19:30:43Z
updatedAt: 2026-11-25T20:23:35Z
lastEditedAt: 2026-11-25T20:23:35Z
closedAt: 2026-11-25T20:04:36Z
---

# status: root: scan the folder of a symlinked working directory

## Problem

In a working directory reached through a symlink, a bare `maintainer status`
exits `1` with `symlink roots are not followed`. The user configured no root, so
the error names a rule the user never invoked.

## PoC

```bash
tmp=$(mktemp -d) && mkdir "$tmp/real" && ln -s real "$tmp/link"
(cd "$tmp/real" && maintainer status --format=plain; echo "exit=$?")  # exit=0
(cd "$tmp/link" && maintainer status --format=plain; echo "exit=$?")
# today:    "symlink roots are not followed", exit=1
# expected: the rows of $tmp/real, exit=0
```

## Expected behavior

A bare `status` in a working directory reached through a symlink scans the
folder that the symlink names and prints the same rows as in that folder.

## Acceptance criteria

- [x] In the PoC, `status` in `link` exits `0` and prints the same rows as in
      `real`.
- [x] A test covers the implicit root through a symlink.
- [x] A root that the user names explicitly is still refused when it is a
      symlink, and its test passes.

## Out of scope

The rule for explicit roots. Symlinks inside the workspace during the scan.

Source: [20261125T152232Z](../reviews/20261125T152232Z.md), "The implicit root fails in a working directory reached through a symlink".
