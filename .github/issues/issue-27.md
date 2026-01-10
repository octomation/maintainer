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
updatedAt: 2022-06-15T10:15:47Z
lastEditedAt:
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

**Current state:** the task is closed and the command is implemented. It accepts one year and uses the current one when the argument is omitted; the `--format` flag from the original prototype does not exist and the output is always JSON. This is a snapshot of the calendar, not a backup of commits or issues.

Comparing snapshots is [#28](issue-28.md). Several years in a single call remain task [#77](issue-77.md), and the daily scenario is [#78](issue-78.md).

<!-- 2022-06-02T19:46Z https://github.com/octomation/maintainer/issues/27#issuecomment-1145276220
https://github.com/octomation/maintainer/commit/f80f8f94c822db0c81049a54f3c53e6cb90e404d
-->
