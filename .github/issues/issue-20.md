---
id: 20
database_id: 981905270
node_id: MDU6SXNzdWU5ODE5MDUyNzA=
status: closed
title: "contribution chart"
labels: ["scope: code"]
url: https://github.com/octomation/maintainer/issues/20
created_at: 2021-08-28T18:57:52Z
updated_at: 2022-05-10T07:21:07Z
---

# contribution chart

Bring the GitHub Contributions Calendar into the terminal: view the activity and get a date for the next contribution. The motivation is to stop studying the calendar in a browser before using a personal Git wrapper.

In the current CLI the scenario reads:

```bash
maintainer github contribution lookup 2021-02-01/3
maintainer github contribution suggest --short 2021/+10
```

The first command shows the calendar, the second returns the suggested date. The `git at` call from the original idea is an external user command, not part of maintainer; `suggest` does not create a commit itself.

**Current state:** the historical task is closed and the features are available under the `github contribution` group. Choosing the date in detail is [#24](issue-24.md); viewing, snapshots, comparison and the activity distribution are [#26](issue-26.md)–[#29](issue-29.md). The scenarios are described in [the documentation](../../docs/changelog.md).
