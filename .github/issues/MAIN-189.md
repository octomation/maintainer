---
code: MAIN-189
id: null
databaseId: null
number: null
url: null
title: "fetch: state: refuse an unknown profile in state prune"
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
createdAt: null
updatedAt: null
lastEditedAt: null
closedAt: null
---

# fetch: state: refuse an unknown profile in state prune

## Problem

`fetch state prune --profile <name>` with a profile that the configuration does
not have prints `pruned 0 record(s)` and exits `0`. The same flag on `fetch`
is a usage error with exit `2`. A typo in the profile name looks like a prune
that found nothing to remove.

## PoC

```bash
maintainer fetch state prune --profile wrok     # the config has "work"
# today:    pruned 0 record(s), exit 0
# expected: a usage error that names "wrok", exit 2, no record removed
```

## Expected behavior

`state prune` checks `--profile` and `--owner` against the configuration as
`fetch` does: an unknown profile is a usage error with exit `2`, and the state
file stays as it is.

## Acceptance criteria

- [ ] A test runs `state prune --profile` with an unknown profile: exit `2`, the
      message names the profile, the state file is unchanged.
- [ ] `fetch --profile` and `state prune --profile` give the same error for the
      same unknown profile.

## Out of scope

Owners that the configuration does not list: `--owner` may name an owner that
is only in the state.

Related: [[MAIN-175]].
