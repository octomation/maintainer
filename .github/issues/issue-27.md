---
code:
id: I_kwDOE2M9Zc5JOiV-
databaseId: 1228547454
number: 27
url: https://github.com/octomation/maintainer/issues/27
title: "implement github contribution snapshot command"
labels:
  - "scope: code"
milestone: "[[milestone-1]]"
state: CLOSED
stateReason: COMPLETED
createdAt: 2022-05-07T06:50:24Z
updatedAt: 2026-09-25T05:11:28Z
lastEditedAt: 2026-09-25T05:11:28Z
closedAt: 2022-06-02T19:46:00Z
---

# implement github contribution snapshot command

Store the GitHub Contributions Calendar for a year in a machine-readable form. A snapshot is needed to compare activity taken at different moments, to archive it, and to process it with other tools.

The working interface:

```bash
maintainer github contribution snapshot 2021 > snapshot.2021.json
```

The result is a JSON object mapping a UTC date to the number of contributions, for example:

```json
{"2021-02-01T00:00:00Z": 5}
```

This is a snapshot of the calendar, not a backup of commits or issues. Comparing snapshots is [[issue-28]]. Several years in a single call remain task [[issue-77]], and the daily scenario is [[issue-78]].

<!-- 2022-06-02T19:46Z https://github.com/octomation/maintainer/issues/27#issuecomment-1145276220
https://github.com/octomation/maintainer/commit/f80f8f94c822db0c81049a54f3c53e6cb90e404d
-->
