---
code:
id: MDU6SXNzdWU5ODE5MDUyNzA=
databaseId: 981905270
number: 20
url: https://github.com/octomation/maintainer/issues/20
title: "contribution chart"
labels:
  - "scope: code"
milestone:
state: CLOSED
stateReason: COMPLETED
createdAt: 2021-08-28T18:57:52Z
updatedAt: 2026-09-25T05:11:06Z
lastEditedAt: 2026-09-25T05:11:06Z
closedAt: 2022-05-10T07:21:07Z
---

# contribution chart

Bring the GitHub Contributions Calendar into the terminal: view the activity and get a date for the next contribution. The motivation is to stop studying the calendar in a browser before using a personal Git wrapper.

The scenario reads:

```bash
maintainer github contribution lookup 2021-02-01/3
maintainer github contribution suggest --short 2021/+10
```

The first command shows the calendar, the second returns the suggested date. The `git at` call from the original idea is an external user command, not part of maintainer; `suggest` does not create a commit itself.

Related: choosing the date in detail [[issue-24]]; viewing, snapshots, comparison and the activity distribution [[issue-26]], [[issue-27]], [[issue-28]], [[issue-29]]; the scenarios in [the documentation](../../docs/content/changelog/index.md).

<!-- 2021-08-28T19:05Z https://github.com/octomation/maintainer/issues/20#issuecomment-907674589
research:
- https://github.com/sallar/github-contributions-chart
- https://github.com/IonicaBizau/git-stats
-->

<!-- 2022-05-10T07:21Z https://github.com/octomation/maintainer/issues/20#issuecomment-1122022184
duplicate #24
-->
