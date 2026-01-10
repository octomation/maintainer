---
code:
id: I_kwDOE2M9Zc5Og3Jn
databaseId: 1317237351
number: 74
url: https://github.com/octomation/maintainer/issues/74
title: "internal: pkg: remove golden package"
labels:
  - "scope: code"
milestone: "[[milestone-1]]"
state: CLOSED
stateReason: COMPLETED
createdAt: 2022-07-25T18:57:39Z
updatedAt: 2022-07-25T19:00:16Z
lastEditedAt:
closedAt: 2022-07-25T19:00:16Z
---

# internal: pkg: remove golden package

Remove the duplicating golden helper area now that there is a shared mechanism for formats and for contribution snapshot sources. This reduces the number of ways to solve one task and the risk of behaviour diverging when test data is handled.

The observable result: the previous scenarios of reading reference data and comparing snapshots keep working through the shared facilities; file formats and diagnostic messages do not change by accident because an old layer was dropped.

**Current state:** the issue is closed and there is no golden package in the current tree. Its responsibilities are split between [pkg/io](../../internal/pkg/io/) and the [contributions sources](../../internal/model/github/contribution/source.go). The related JSON/YAML work is [#71](issue-71.md). This refactoring requires no user-facing flag and no new CLI command.
