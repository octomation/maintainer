---
code:
id: I_kwDOE2M9Zc5LHtsg
databaseId: 1260313376
number: 44
url: https://github.com/octomation/maintainer/issues/44
title: "github: contribution: refactor command and view"
labels:
  - "scope: code"
  - "scope: test"
  - "type: improvement"
  - "impact: medium"
  - "effort: medium"
milestone: "[[milestone-1]]"
state: OPEN
stateReason:
createdAt: 2022-06-03T19:53:46Z
updatedAt: 2023-04-06T11:16:42Z
lastEditedAt:
closedAt:
---

# github: contribution: refactor command and view

Make the contribution commands easier to maintain before the release: a change to the range calculation or to the rendering must not break neighbouring scenarios. The task began as technical-debt work ahead of v0.1.0, but its outcome is stable, consistent CLI behaviour.

The reference user path: `lookup` shows the dates → `suggest` highlights the chosen day → `snapshot` stores the data → `diff` explains the changes. Identical dates and counts must be interpreted identically in every view.

**Current state:** the commands are already split across files and share the date and table operations. At the same time `diff` keeps a separate view carrying TODOs, and the table prints `Stats: coming soon`. Neither treating the whole original technical debt as untouched, nor treating it as finished, is accurate.

The remaining work should be aligned with [#127](issue-127.md), the diff fix [#70](issue-70.md) and the statistics [#79](issue-79.md). The completion criterion is preserved scenarios and specific divergences eliminated, not a reorganization of files as an end in itself.
