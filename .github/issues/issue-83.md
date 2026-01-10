---
code:
id: I_kwDOE2M9Zc5Qpzfj
databaseId: 1353136099
number: 83
url: https://github.com/octomation/maintainer/issues/83
title: "deps: up go-github top v47.0.0"
labels:
  - "scope: code"
milestone: "[[milestone-1]]"
state: CLOSED
stateReason: COMPLETED
createdAt: 2022-08-27T20:21:29Z
updatedAt: 2026-09-25T05:14:00Z
lastEditedAt: 2026-09-25T05:14:00Z
closedAt: 2022-09-20T12:34:13Z
---

# deps: up go-github top v47.0.0

Update the GitHub client to v47.0.0 while keeping maintainer's user scenarios working. This is routine dependency maintenance: take upstream changes without accumulating a large gap between versions.

The original target release: [v47.0.0](https://github.com/google/go-github/releases/tag/v47.0.0). New client capabilities must not be counted as new maintainer features: they only become available once a command uses them.

The update is done when the project builds, authorization and GitHub data retrieval keep their previous contract, and the checks for the affected scenarios pass. Crossing a major version also requires aligning module and import paths; editing a single version number is not enough.
