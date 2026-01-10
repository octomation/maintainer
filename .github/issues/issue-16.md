---
code:
id: MDU6SXNzdWU4NDEwNzkwNTY=
databaseId: 841079056
number: 16
url: https://github.com/octomation/maintainer/issues/16
title: "org: prepare release v0.2.0"
labels:
  - "scope: docs"
  - "scope: code"
  - "scope: test"
milestone:
state: CLOSED
stateReason: NOT_PLANNED
createdAt: 2021-03-25T16:03:25Z
updatedAt: 2023-03-31T15:34:20Z
lastEditedAt:
closedAt: 2023-03-31T15:32:18Z
---

# org: prepare release v0.2.0

Prepare the GitHub Labels direction for the v0.2.0 release: make the user scenarios clear and verify the label operations before distributing the tool.

The original scope covered clearing technical debt, checking code quality, writing tests, and documenting the `github labels` commands. The release scenario for a user: read the current set → choose a preset → review the changes → apply → confirm that a second run changes nothing.

Readiness for such a release means the help and the examples agree with the available CLI, that adding, changing and removing labels are all verified, and that authorization and malformed-input errors are understandable.

**Current state:** the historical task is closed. The v0.2.0 number belongs to the former milestone plan ([#47](issue-47.md)); labels were later dropped from the product ([#80](issue-80.md)). This is neither a live promise about a future release nor a reason to document removed commands as available.
