---
code: MAIN-179
id: null
databaseId: null
number: null
url: null
title: "fetch: verbosity: make each offered level add output"
issueType: Bug
labels:
  - "scope: docs"
  - "scope: code"
  - "scope: test"
  - "type: bug"
  - "severity: minor"
  - "impact: low"
  - "effort: easy"
assignees:
  - kamilsk
milestone: null
parent: null
issueFields: {}
state: OPEN
stateReason: null
duplicateOf: null
createdAt: null
updatedAt: null
lastEditedAt: null
closedAt: null
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

- [ ] A test runs one apply at `-v`, `-vv` and `-vvv`: each offered level prints
      lines that the level below does not.
- [ ] `fetch --help` and `docs/content/fetch.md` list the same levels and say
      what each adds.

## Out of scope

`-q` and the format of the lines.

Source: [20261125T152232Z](../reviews/20261125T152232Z.md), "The `-vv` and `-vvv` levels print nothing beyond `-v`".
