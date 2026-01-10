---
code:
id: I_kwDOE2M9Zc5UUIkX
databaseId: 1414564119
number: 86
url: https://github.com/octomation/maintainer/issues/86
title: "deps: up go-github to v48.0.0"
labels:
  - "scope: code"
milestone:
state: CLOSED
stateReason: COMPLETED
createdAt: 2022-10-19T08:48:40Z
updatedAt: 2022-11-15T08:37:44Z
lastEditedAt:
closedAt: 2022-11-15T08:37:44Z
---

# deps: up go-github to v48.0.0

Update the GitHub client to v48.0.0 while keeping maintainer's user scenarios working. This is routine dependency maintenance: take upstream changes without accumulating a large gap between versions.

The original target release: [v48.0.0](https://github.com/google/go-github/releases/tag/v48.0.0). New client capabilities must not be counted as new maintainer features: they only become available once a command uses them.

The update is done when the project builds, authorization and GitHub data retrieval keep their previous contract, and the checks for the affected scenarios pass. Crossing a major version also requires aligning module and import paths; editing a single version number is not enough.

**Current state:** this historical issue is closed. [go.mod](../../go.mod) already uses `github.com/google/go-github/v91 v91.0.0`, so returning the dependency to v48.0.0 is not required. This is the record of one concrete update step, not a request to downgrade.
