---
code: MAIN-180
id: null
databaseId: null
number: null
url: null
title: "fetch: discovery: discover the owners listed next to the wildcard"
issueType: Bug
labels:
  - "scope: docs"
  - "scope: code"
  - "scope: test"
  - "type: bug"
  - "severity: minor"
  - "impact: medium"
  - "effort: easy"
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

# fetch: discovery: discover the owners listed next to the wildcard

## Problem

With `include_owners = ["*", "stranger"]`, discovery covers only the user and
the member organizations. The allowlist then drops the repositories of
`stranger` without a word. The fetch guide advises listing owners outside the
member organizations next to the wildcard.

## PoC

```bash
# fetch.toml: include_owners = ["*", "<an owner outside your organizations>"]
maintainer fetch
# today:    no repository of that owner in the plan, no message
# expected: the repositories of that owner are planned
```

## Expected behavior

A list that mixes `*` with named owners discovers the user, the member
organizations and each named owner.

## Acceptance criteria

- [ ] A discovery test with `["*", "stranger"]` requests the repositories of
      `stranger` and plans them next to those of the user.
- [ ] `docs/content/fetch.md` says what such a list discovers.

## Out of scope

The meaning of a list without `*`.

Related: [[MAIN-133]].

Source: [20261125T152232Z](../reviews/20261125T152232Z.md), "An owner listed next to the wildcard in `include_owners` is silently dropped".
