---
code: MAIN-176
id: I_kwDOE2M9Zc8AAAABih0ErA
databaseId: 6612124844
number: 559
url: https://github.com/octomation/maintainer/issues/559
title: "github: contribution: print the lookup table on stdout"
issueType: Bug
labels:
  - "effort: easy"
  - "impact: medium"
  - "scope: code"
  - "scope: test"
  - "severity: minor"
  - "type: bug"
assignees:
  - kamilsk
milestone: null
parent: null
issueFields: {}
state: OPEN
stateReason: null
duplicateOf: null
createdAt: 2026-11-25T19:31:19Z
updatedAt: 2026-11-25T19:31:20Z
lastEditedAt: null
closedAt: null
---

# github: contribution: print the lookup table on stdout

## Problem

`maintainer github contribution lookup > week.txt` writes an empty file, and a
pipe to `grep` gets nothing. The calendar table, the only result of `lookup`,
goes to stderr.

## PoC

```bash
tmp=$(mktemp -d)
maintainer github contribution lookup > "$tmp/week.txt"; wc -c < "$tmp/week.txt"
# today:    0
# expected: the size of the table
```

## Expected behavior

`lookup` prints its table on stdout. Diagnostics stay on stderr, and `suggest`
keeps its table on stderr.

## Acceptance criteria

- [x] The PoC writes the table to the file, and stderr holds only diagnostics.
- [x] A command test captures stdout and stderr of `lookup` and `suggest`
      separately and checks where each table goes.

## Out of scope

The output of `histogram`.

Source: [20261125T152232Z](../reviews/20261125T152232Z.md), "`lookup` prints its calendar table to stderr, so stdout stays empty".
