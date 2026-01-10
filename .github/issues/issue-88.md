---
code:
id: I_kwDOE2M9Zc5ZfG7a
databaseId: 1501327066
number: 88
url: https://github.com/octomation/maintainer/issues/88
title: "git: stats: provide local-based statistics"
labels:
  - "scope: code"
milestone:
state: OPEN
stateReason:
createdAt: 2022-12-17T10:58:56Z
updatedAt: 2026-09-25T05:14:13Z
lastEditedAt: 2026-09-25T05:14:13Z
closedAt:
---

# git: stats: provide local-based statistics

Derive statistics directly from the local Git history, to get insights the profile calendar cannot give. This is useful for private projects, for working without GitHub, and for analysing individual repositories that do not match the aggregated profile calendar.

A provisional interface:

```bash
maintainer git stats --since=2022-01-01
# The period, the authors, the number of commits and the distribution by day.
```

The scope of history and of authors has to be defined, along with the treatment of merge commits and of the time zone. The result must state explicitly which local history it counted; the absence of a network request does not automatically make statistics for all branches or remote repositories complete.

The checkout status table from the status specification is a separate task, not history statistics.

Related: [git-stats](https://github.com/IonicaBizau/git-stats), [github-contributions-chart](https://github.com/sallar/github-contributions-chart).
