---
code:
id: I_kwDOE2M9Zc5wQ1wQ
databaseId: 1883462672
number: 158
url: https://github.com/octomation/maintainer/issues/158
title: "github: up google/go-github to v55.0.0"
labels:
  - "type: improvement"
  - "scope: deps"
  - "impact: low"
  - "effort: easy"
milestone: "[[milestone-1]]"
state: CLOSED
stateReason: COMPLETED
createdAt: 2023-09-06T07:56:34Z
updatedAt: 2023-11-16T09:21:10Z
lastEditedAt:
closedAt: 2023-11-16T09:21:10Z
---

# github: up google/go-github to v55.0.0

Update the GitHub client to v55.0.0 while keeping maintainer's user scenarios working. This is routine dependency maintenance: take upstream changes without accumulating a large gap between versions.

The original target release: [v55.0.0](https://github.com/google/go-github/releases/tag/v55.0.0). New client capabilities must not be counted as new maintainer features: they only become available once a command uses them.

The update is done when the project builds, authorization and GitHub data retrieval keep their previous contract, and the checks for the affected scenarios pass. Crossing a major version also requires aligning module and import paths; editing a single version number is not enough.

**Current state:** this historical issue is closed. [go.mod](../../go.mod) already uses `github.com/google/go-github/v91 v91.0.0`, so returning the dependency to v55.0.0 is not required. This is the record of one concrete update step, not a request to downgrade.

<!-- 2023-09-06T08:00Z https://github.com/octomation/maintainer/issues/158#issuecomment-1707858989
need to be adopted
- https://github.com/google/go-github/pull/2895
- https://github.com/octomation/maintainer/blob/main/internal/pkg/http/client.go
-->

<!-- 2023-11-16T09:21Z https://github.com/octomation/maintainer/issues/158#issuecomment-1814067194
already fixed
-->
