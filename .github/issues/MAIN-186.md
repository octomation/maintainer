---
code: MAIN-186
id: I_kwDOE2M9Zc8AAAABige75A
databaseId: 6610729956
number: 547
url: https://github.com/octomation/maintainer/issues/547
title: "tools: env: keep a stored token for one account"
issueType: Bug
labels:
  - "scope: code"
  - "scope: test"
  - "type: bug"
  - "severity: major"
  - "impact: high"
  - "effort: easy"
assignees:
  - kamilsk
milestone: null
parent: null
issueFields: {}
state: OPEN
stateReason: null
duplicateOf: null
createdAt: 2026-11-25T18:09:57Z
updatedAt: 2026-11-25T18:12:44Z
lastEditedAt: 2026-11-25T18:10:35Z
closedAt: null
---

# tools: env: keep a stored token for one account

## Problem

The repository helper that stores a GitHub token for `gh` leaves it readable by
other accounts of the machine. The stored token must stay readable by its owner
only, and the value must not appear in any process argument.

## Expected behavior

- `@env set` creates `.env` with mode `0600` and sets that mode on an existing
  file before it appends a key.
- An update rewrites the file with shell builtins only: read, replace the line,
  write a new temporary file created exclusively with mode `0600` in the same
  folder, rename. The value is in no process argument.

## Acceptance criteria

- [x] After `@env set KEY value` in a scratch folder, `.env` has mode `0600`
      for a new file, an appended key and a replaced key.
- [x] A replaced key leaves the other lines as they were; a value with `|`, `&`
      or `/` is stored as given.
- [x] `shellcheck` reports nothing new for `bin/lib/utils/env.bash`.

## Out of scope

- The product does not read this file: [[MAIN-177]].
- No changelog entry: this is a tool of the maintainer, not a part of the
  product.

Source: [the security review of 2026-11-25](../reviews/20261125T152238Z.md).
