---
code:
id: I_kwDOE2M9Zc5h2PpZ
databaseId: 1641609817
number: 121
url: https://github.com/octomation/maintainer/issues/121
title: "github: contribution: lookup command is corrupted"
labels:
  - "scope: code"
  - "scope: test"
  - "type: bug"
  - "severity: major"
  - "impact: medium"
  - "effort: easy"
milestone: "[[milestone-1]]"
state: CLOSED
stateReason: COMPLETED
createdAt: 2023-03-27T08:09:12Z
updatedAt: 2026-09-25T05:09:50Z
lastEditedAt: 2026-09-25T05:09:50Z
closedAt: 2023-03-29T13:41:39Z
---

# github: contribution: lookup command is corrupted

Restore support for a short date in lookup: a year with a direction must not be rejected as an incomplete day notation. The rejection made exploring a long period awkward.

The reproduction for [v0.1.0-rc10](https://github.com/octomation/maintainer/releases/tag/v0.1.0-rc10):

```bash
maintainer github contribution lookup 2022/+10
# Error: parsing "2022" as YYYY-MM-DD.
```

The year is expected to be accepted as the reference date and a directed weekly window built from it. A short year input with `/+10` does not mean showing every day of that year.

Related: a bare `lookup 2022` without a suffix [[issue-155]].
