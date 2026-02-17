---
code: MAIN-169
id: I_kwDOE2M9Zc8AAAABihzk1g
databaseId: 6612116694
number: 553
url: https://github.com/octomation/maintainer/issues/553
title: "service: hub: keep the last good version of an issue after a refused triage"
issueType: Bug
labels:
  - "effort: medium"
  - "impact: medium"
  - "scope: code"
  - "scope: test"
  - "severity: major"
  - "type: bug"
assignees:
  - kamilsk
milestone: null
parent: null
issueFields: {}
state: CLOSED
stateReason: COMPLETED
duplicateOf: null
createdAt: 2026-11-25T19:30:52Z
updatedAt: 2026-11-25T20:23:40Z
lastEditedAt: 2026-11-25T20:23:40Z
closedAt: 2026-11-25T20:07:04Z
---

# service: hub: keep the last good version of an issue after a refused triage

## Problem

An issue whose file has a broken frontmatter disappears from the hub when the
user changes one of its properties. The hub answers `412` and removes the issue
from the index; a reload gives `404`, and **Reindex** does not bring it back.
`docs/content/service.md` promises that such a file keeps its last good
version.

## PoC

1. Open an issue in `/inbox`, then break the frontmatter of its file in an
   editor, for example with `broken: [`.
2. Set a label of that issue in the hub.
3. Today: `412`, the issue leaves the list, a reload gives `404`, and
   **Reindex** does not return it. Expected: `412`, and the issue stays in the
   list with its last good fields, marked stale.

## Expected behavior

When the hub refuses a triage write because the file no longer parses, it keeps
the last good version of the issue, flagged stale, as a scan does. The file stays
as the user left it. A scan after the user repairs the file shows the new
version.

## Acceptance criteria

- [x] A test breaks the frontmatter of an indexed issue and patches a label with
      the old version: the answer is `412`, and the index keeps the issue,
      flagged stale, with its last good fields.
- [x] A scan after that refusal still holds the stale entry.
- [x] The file is unchanged byte for byte after the refused write.

## Out of scope

The repair of the file. The text of the refusal.

Source: [20261125T152232Z](../reviews/20261125T152232Z.md), "A refused triage on a malformed file drops the issue instead of keeping it stale".
