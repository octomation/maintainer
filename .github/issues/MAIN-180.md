---
code: MAIN-180
id: I_kwDOE2M9Zc8AAAABih0UvQ
databaseId: 6612128957
number: 562
url: https://github.com/octomation/maintainer/issues/562
title: "fetch: discovery: discover the owners listed next to the wildcard"
issueType: Bug
labels:
  - "effort: easy"
  - "impact: medium"
  - "scope: code"
  - "scope: docs"
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
createdAt: 2026-11-25T19:31:33Z
updatedAt: 2026-11-25T20:24:05Z
lastEditedAt: 2026-11-25T20:24:05Z
closedAt: 2026-11-25T20:03:03Z
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

- [x] A discovery test with `["*", "stranger"]` requests the repositories of
      `stranger` and plans them next to those of the user.
- [x] `docs/content/fetch.md` says what such a list discovers.

## Out of scope

The meaning of a list without `*`.

Related: [[MAIN-133]].

Source: [20261125T152232Z](../reports/20261125T152232Z.md), "An owner listed next to the wildcard in `include_owners` is silently dropped".
