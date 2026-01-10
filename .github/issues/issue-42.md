---
code:
id: I_kwDOE2M9Zc5KbHqg
databaseId: 1248623264
number: 42
url: https://github.com/octomation/maintainer/issues/42
title: "deps: up go-github to v45.0.0"
labels:
  - "scope: code"
milestone:
state: CLOSED
stateReason: COMPLETED
createdAt: 2022-05-25T19:38:59Z
updatedAt: 2026-09-25T05:12:11Z
lastEditedAt: 2026-09-25T05:12:11Z
closedAt: 2022-06-02T19:17:18Z
---

# deps: up go-github to v45.0.0

Update the GitHub client to v45.0.0 while keeping maintainer's user scenarios working. This is routine dependency maintenance: take upstream changes without accumulating a large gap between versions.

The original target release: [v45.0.0](https://github.com/google/go-github/releases/tag/v45.0.0). New client capabilities must not be counted as new maintainer features: they only become available once a command uses them.

The update is done when the project builds, authorization and GitHub data retrieval keep their previous contract, and the checks for the affected scenarios pass. Crossing a major version also requires aligning module and import paths; editing a single version number is not enough.
