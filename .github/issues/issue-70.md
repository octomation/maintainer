---
code:
id: I_kwDOE2M9Zc5OazM6
databaseId: 1315648314
number: 70
url: https://github.com/octomation/maintainer/issues/70
title: "github: contribution: refactor diff command"
labels:
  - "scope: docs"
  - "scope: code"
  - "scope: test"
  - "type: improvement"
  - "impact: medium"
  - "effort: medium"
milestone: "[[milestone-1]]"
state: OPEN
stateReason:
createdAt: 2022-07-23T12:31:24Z
updatedAt: 2023-04-06T11:15:57Z
lastEditedAt:
closedAt:
---

# github: contribution: refactor diff command

Fix the meaning and the presentation of a calendar diff. When contributions grew, the command printed negative values, so the user could not tell what had been added since the snapshot was stored. The original report also asked for the two sources to become positional arguments, following the conventions of [the `diff` command](https://linuxize.com/post/diff-command-in-linux/).

The working interface already uses positional arguments:

```bash
maintainer github contribution diff before.json after.json
```

The contract to verify:

```text
Day          before  after  diff
2013-11-13      1      5    +4
2013-11-14      5      2    -3
```

Dates present in only one of the snapshots must be accounted for, as must changes at the Sunday and year boundaries; identical snapshots must report explicitly that there is no diff.

**Confirmed against the current CLI:** on two local JSON snapshots with the dates and values above, growth prints as `+4`, while a decrease prints as `+18446744073709551613` instead of `-3`. The difference is stored in an unsigned counter; the problem reproduces without GitHub and does not depend on the calendar being updated.

The move from `--base`/`--head` to arguments is already done. What remains is correct comparison and presentation, including decreasing activity. The original context is [1522818](https://github.com/octomation/maintainer/commit/1522818c8dba3194378ffb83088be880b0245c32); the accompanying work is [#127](issue-127.md).
