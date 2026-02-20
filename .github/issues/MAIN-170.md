---
code: MAIN-170
id: I_kwDOE2M9Zc8AAAABi07sbA
databaseId: 6632172652
number: 571
url: https://github.com/octomation/maintainer/issues/571
title: "status: inspection: keep one slow checkout from holding the whole run"
issueType: Bug
labels:
  - "scope: docs"
  - "scope: code"
  - "scope: test"
  - "type: bug"
  - "severity: minor"
  - "impact: medium"
  - "effort: medium"
assignees:
  - kamilsk
milestone: null
parent: null
issueFields: {}
state: CLOSED
stateReason: COMPLETED
duplicateOf: null
createdAt: 2026-11-26T17:42:00Z
updatedAt: 2026-11-26T20:01:40Z
lastEditedAt: 2026-11-26T20:01:40Z
closedAt: 2026-11-26T19:51:41Z
---

# status: inspection: keep one slow checkout from holding the whole run

## Problem

On 2026-10-19 and 2026-10-20 a `status` run over 164 checkouts took about 4
minutes. About 2.5 of them went to `git diff --find-renames --numstat HEAD` in
one checkout, where the same command took about 2 s in September. The user waits
for the whole snapshot because of one row, and the hub `/status` waits too.

## PoC

```bash
time maintainer status --format=plain >/dev/null
# with one checkout whose rename detection takes minutes:
# today:    the run takes as long as that checkout
# expected: the run takes about the time of the others plus a fixed bound
```

## Expected behavior

- One slow checkout does not set the time of a `status` run: the run finishes
  in about the time of the other checkouts plus a fixed bound.
- The row of a slow checkout keeps its branch, divergence and push lock. Counts
  that were not computed show as unknown, never as zero.
- The plain, JSON and hub views show such a row the same way.

## Acceptance criteria

- [x] A test with a diff that does not return finishes within the bound, with
      the row present and its counts unknown.
- [x] With a checkout whose rename detection takes about 25 s in `git
      status` and in `git diff`, such as one with 900 staged renames of large
      files, `time maintainer status --format=plain` takes about the bound, not
      their sum; on the owner's workspace, where the slow checkout of
      2026-10-19 is gone, the run takes as long as before the change.
- [x] `docs/content/status.md` shows how a row with unknown counts looks.

## Out of scope

How renames are counted: [[MAIN-130]].

Source: [20261125T152232Z](../reviews/20261125T152232Z.md), "Rename detection in one checkout holds the whole `status` run for minutes".
