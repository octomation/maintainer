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
updatedAt: 2026-09-25T05:12:16Z
lastEditedAt: 2026-09-25T05:12:16Z
closedAt:
---

# github: contribution: refactor command and view

Make the contribution commands easier to maintain before the release: a change to the range calculation or to the rendering must not break neighbouring scenarios. The task began as technical-debt work ahead of v0.1.0, but its outcome is stable, consistent CLI behaviour.

The reference user path: `lookup` shows the dates → `suggest` highlights the chosen day → `snapshot` stores the data → `diff` explains the changes. Identical dates and counts must be interpreted identically in every view.

The remaining work should be aligned with [[issue-127]], the diff fix [[issue-70]] and the statistics [[issue-79]]. The completion criterion is preserved scenarios and specific divergences eliminated, not a reorganization of files as an end in itself.
