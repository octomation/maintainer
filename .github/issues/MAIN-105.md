---
code: MAIN-105
id: I_kwDOE2M9Zc5jGx7W
databaseId: 1662721750
number: 132
url: https://github.com/octomation/maintainer/issues/132
title: "linear: generate client"
issueType: null
labels:
  - "scope: code"
  - "type: feature"
  - "scope: deps"
  - "impact: high"
  - "effort: medium"
assignees:
  - kamilsk
milestone: "[[MAIN-M3]]"
parent: null
issueFields: {}
state: OPEN
stateReason: null
duplicateOf: null
createdAt: 2023-04-11T15:13:16Z
updatedAt: 2026-11-20T15:42:16Z
lastEditedAt: 2026-11-20T15:42:16Z
closedAt: null
---

# linear: generate client

## Motivation

maintainer gives a contributor one picture of the work across trackers, not
only GitHub. Linear is one of them. Today the structure of a Linear project is
seen only by walking between cards in the browser. The user should see it from
the terminal: a flat list of issues, and a tree of parents and sub-issues.

The enabler is a typed client for Linear's GraphQL API, generated with
[genqlient](https://github.com/Khan/genqlient) from the
[Linear schema](https://github.com/linear/linear/blob/master/packages/sdk/src/schema.graphql).
Generating it is a means, not a value in itself: it counts only with data
retrieval that is checked. The old `docs/INTRODUCTION.md` path of genqlient is
no longer reachable, so the project address is kept instead.

## PoC

The interface is provisional:

```bash
maintainer linear issues --flat
maintainer linear issues --tree
```

## Acceptance criteria

- [ ] Both views show the identifier, the title and the status of each issue,
      and the same set of issues.
- [ ] The tree keeps the parent and sub-issue relations, and does not lose an
      issue that has no parent.
- [ ] An access error and a partial load are visible, never a silently shorter
      list.
- [ ] The client is generated from a pinned schema by `go generate`, and
      regenerating it with no schema change produces no diff.
- [ ] The token comes from the configuration or the environment and is never
      printed.

## Out of scope

- Two-way synchronization with Linear.
- Showing Linear issues in the local hub next to the GitHub mirror.
- Other trackers, such as Jira.

<!-- 2026-11-20T21:00Z https://github.com/octomation/maintainer/issues/132#issuecomment-7043711584
Maturity: idea. There is no Linear code: no client, no command, no genqlient dependency. The nearest code reads GitHub issues: the mirror reader of the hub (`internal/service/hub/mirror`) and the unwired `github management issue` (`internal/command/github/management/issue.go`).
v0.5.0 proposal: keep, outside the v0.5.0 scope. Bringing the issues of another tracker into the contributor's one picture is the mission itself; v0.5.0 removes nothing it would need.
-->
