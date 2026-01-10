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
updatedAt: 2026-09-25T05:13:26Z
lastEditedAt: 2026-09-25T05:13:26Z
closedAt:
---

# github: contribution: refactor diff command

Fix the meaning and the presentation of a calendar diff. When contributions grew, the command printed negative values, so the user could not tell what had been added since the snapshot was stored. The original report also asked for the two sources to become positional arguments, following the conventions of [the `diff` command](https://linuxize.com/post/diff-command-in-linux/).

The expected interface uses positional arguments:

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

Related: the original context [1522818](https://github.com/octomation/maintainer/commit/1522818c8dba3194378ffb83088be880b0245c32), the accompanying work [[issue-127]].
