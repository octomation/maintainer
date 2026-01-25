---
code: MAIN-74
id: I_kwDOE2M9Zc5Og3Jn
databaseId: 1317237351
number: 74
url: https://github.com/octomation/maintainer/issues/74
title: "internal: pkg: remove golden package"
issueType: null
labels:
  - "scope: code"
assignees:
  - kamilsk
milestone: "[[MAIN-M1]]"
parent: null
issueFields: {}
state: CLOSED
stateReason: COMPLETED
duplicateOf: null
createdAt: 2022-07-25T18:57:39Z
updatedAt: 2026-09-25T05:13:36Z
lastEditedAt: 2026-09-25T05:13:36Z
closedAt: 2022-07-25T19:00:16Z
---

# internal: pkg: remove golden package

Remove the duplicating golden helper area now that there is a shared mechanism for formats and for contribution snapshot sources. This reduces the number of ways to solve one task and the risk of behaviour diverging when test data is handled.

The observable result: the previous scenarios of reading reference data and comparing snapshots keep working through the shared facilities; file formats and diagnostic messages do not change by accident because an old layer was dropped.

This refactoring requires no user-facing flag and no new CLI command.

Related: the JSON/YAML work [[MAIN-71]].
