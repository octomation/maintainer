---
code: MAIN-198
id: I_kwDOE2M9Zc8AAAABi08Bwg
databaseId: 6632178114
number: 576
url: https://github.com/octomation/maintainer/issues/576
title: "fetch: plan: name the path of a scan conflict that has no repository"
issueType: Bug
labels:
  - "effort: easy"
  - "impact: low"
  - "scope: code"
  - "scope: test"
  - "severity: minor"
  - "type: bug"
assignees:
  - kamilsk
milestone: null
parent: null
issueFields: {}
state: OPEN
stateReason: null
duplicateOf: null
createdAt: 2026-11-26T17:42:22Z
updatedAt: 2026-11-26T17:43:05Z
lastEditedAt: 2026-11-26T17:43:05Z
closedAt: null
---

# fetch: plan: name the path of a scan conflict that has no repository

## Problem

A directory of the workspace that the scan cannot read becomes a conflict
without a repository identity. The plan prints it as `✗ conflict   /`, and the
total `plan: 1 repos total` leaves the row out, so the user reads a
repository without a name.

## PoC

```bash
chmod 0 "$root/public/acme"
maintainer fetch
# today:    ✗ conflict   /
#             at <root>/public/acme
# expected: the row names <root>/public/acme or the scan, never "/"
```

## Expected behavior

A conflict without a repository identity prints its path, or a label of the
scan, in place of an empty `owner/name`. The plain and JSON output agree on
it.

## Acceptance criteria

- [x] A test with an unreadable directory renders a conflict row that names
      its path and contains no bare `/` as a name.
- [x] The JSON output of the same plan names the path of that conflict.

## Out of scope

Which directories the scan reads: [[MAIN-166]].

Source: [20261125T152232Z](../reviews/20261125T152232Z.md), "A scan obstruction conflict prints `/` as its repository name".
