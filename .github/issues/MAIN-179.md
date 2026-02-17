---
code: MAIN-179
id: I_kwDOE2M9Zc8AAAABih0Pew
databaseId: 6612127611
number: 561
url: https://github.com/octomation/maintainer/issues/561
title: "fetch: verbosity: make each offered level add output"
issueType: Bug
labels:
  - "effort: easy"
  - "impact: low"
  - "scope: code"
  - "scope: docs"
  - "scope: test"
  - "severity: minor"
  - "type: bug"
assignees:
  - kamilsk
milestone: null
parent: null
issueFields: {}
state: CLOSED
stateReason: COMPLETED
duplicateOf: null
createdAt: 2026-11-25T19:31:29Z
updatedAt: 2026-11-25T20:24:03Z
lastEditedAt: 2026-11-25T20:24:03Z
closedAt: 2026-11-25T20:02:14Z
---

# fetch: verbosity: make each offered level add output

## Problem

`fetch -vv` and `fetch -vvv` print the same as `fetch -v`, though `--help` and
`docs/content/fetch.md` offer them to diagnose a run.

## PoC

```bash
for v in -v -vv -vvv; do maintainer fetch "$v" 2>&1 >/dev/null | wc -l; done
# today:    the same count for each level
# expected: each level prints more than the level below
```

## Expected behavior

Each verbosity level that `--help` and the guide offer adds output to the level
below it, and the guide says what each level adds. A level that adds nothing is
not offered.

## Acceptance criteria

- [x] A test runs one apply at `-v`, `-vv` and `-vvv`: each offered level prints
      lines that the level below does not.
- [x] `fetch --help` and `docs/content/fetch.md` list the same levels and say
      what each adds.

## Out of scope

`-q` and the format of the lines.

Source: [20261125T152232Z](../reviews/20261125T152232Z.md), "The `-vv` and `-vvv` levels print nothing beyond `-v`".
