---
code: MAIN-16
id: MDU6SXNzdWU4NDEwNzkwNTY=
databaseId: 841079056
number: 16
url: https://github.com/octomation/maintainer/issues/16
title: "org: prepare release v0.2.0"
issueType: null
labels:
  - "scope: docs"
  - "scope: code"
  - "scope: test"
assignees:
  - kamilsk
milestone: "[[MAIN-M1]]"
parent: null
issueFields: {}
state: CLOSED
stateReason: NOT_PLANNED
duplicateOf: null
createdAt: 2021-03-25T16:03:25Z
updatedAt: 2026-10-20T07:31:38Z
lastEditedAt: 2026-09-25T05:10:42Z
closedAt: 2023-03-31T15:32:18Z
---

# org: prepare release v0.2.0

Prepare the GitHub Labels direction for the v0.2.0 release: make the user scenarios clear and verify the label operations before distributing the tool.

The original scope covered clearing technical debt, checking code quality, writing tests, and documenting the `github labels` commands. The release scenario for a user: read the current set → choose a preset → review the changes → apply → confirm that a second run changes nothing.

Readiness for such a release means the help and the examples agree with the available CLI, that adding, changing and removing labels are all verified, and that authorization and malformed-input errors are understandable.

Related: the former milestone plan [[MAIN-47]], label scope removal [[MAIN-80]].
