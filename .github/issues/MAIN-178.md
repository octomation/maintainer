---
code: MAIN-178
id: I_kwDOE2M9Zc8AAAABih0JqA
databaseId: 6612126120
number: 560
url: https://github.com/octomation/maintainer/issues/560
title: "service: config: accept the same repository matches as service.cue"
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
state: CLOSED
stateReason: COMPLETED
duplicateOf: null
createdAt: 2026-11-25T19:31:24Z
updatedAt: 2026-11-27T18:41:00Z
lastEditedAt: 2026-11-27T18:41:00Z
closedAt: 2026-11-25T20:06:33Z
---

# service: config: accept the same repository matches as service.cue

## Problem

A `service.toml` rule with `match = { id = -5 }` loads, the service starts, and
the rule matches nothing. `docs/content/service.md` says that a bad value stops
the service. A rule with `match = { id = 0, owner = "a", name = "b" }` works in
the service but fails `cue vet` against `internal/config/service.cue`.

## PoC

```bash
# service.toml with a [[repos]] rule: match = { id = -5 }
cue vet -c internal/config/service.cue "$tmp/service.toml"        # refused
maintainer service run --config "$tmp/service.toml"
# today:    the service starts
# expected: the service stops with an error that names the rule and id
```

## Expected behavior

The service loader and `internal/config/service.cue` accept and refuse the same
repository matches. A match with an `id` that is zero or negative stops the
service with an error that names the file, the rule and the key.

## Acceptance criteria

- [x] `service run` with `match = { id = -5 }` or `match = { id = 0, owner =
      "a", name = "b" }` stops with that error.
- [x] A test feeds the same valid and invalid matches to the loader and to the
      schema, and both give the same verdict for each.

Source: [20261125T152232Z](../reports/20261125T152232Z.md), "The service loader accepts repos match ids that service.cue rejects".
