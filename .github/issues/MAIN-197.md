---
code: MAIN-197
id: I_kwDOE2M9Zc8AAAABi079oQ
databaseId: 6632177057
number: 575
url: https://github.com/octomation/maintainer/issues/575
title: "build: dist: keep bin/install when godownloader is missing"
issueType: Bug
labels:
  - "scope: code"
  - "type: bug"
  - "severity: minor"
  - "impact: low"
  - "effort: easy"
assignees:
  - kamilsk
milestone: null
parent: null
issueFields: {}
state: CLOSED
stateReason: COMPLETED
duplicateOf: null
createdAt: 2026-11-26T17:42:18Z
updatedAt: 2026-11-26T20:01:55Z
lastEditedAt: 2026-11-26T20:01:55Z
closedAt: 2026-11-26T19:46:38Z
---

# build: dist: keep bin/install when godownloader is missing

## Problem

`make go-dist-installer` runs `godownloader .goreleaser.yml >bin/install`. The
shell truncates `bin/install` before it starts the command, and `godownloader`
is not a tool of `tools/go.mod`. On a machine without it the target fails with
`godownloader: command not found` and leaves an empty `bin/install`, the
tracked install script that users pipe to `sh`.

## PoC

```bash
make go-dist-installer      # godownloader is not installed
# today:    command not found, bin/install has 0 bytes
# expected: the target fails and bin/install is unchanged
```

## Expected behavior

No target overwrites `bin/install` with the output of a tool that the
repository does not install. When the target cannot make the script, it fails
and leaves the tracked file as it is.

## Acceptance criteria

- [x] Without `godownloader`, `make go-dist-installer` exits non-zero and
      `git diff --exit-code bin/install` passes.
- [x] With the tool, the target writes the script only after the tool
      succeeds, or the target is removed together with its mentions.

## Out of scope

The content of the install script.

Source: [20261125T152232Z](../reports/20261125T152232Z.md), "`make go-dist-installer` empties the tracked `bin/install`".
