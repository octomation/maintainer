---
code:
id: I_kwDOE2M9Zc5LUun7
databaseId: 1263725051
number: 49
url: https://github.com/octomation/maintainer/issues/49
title: "deps: up go-github to v45.1.0"
labels:
  - "scope: code"
milestone:
state: CLOSED
stateReason: COMPLETED
createdAt: 2022-06-07T18:50:36Z
updatedAt: 2026-09-25T05:12:28Z
lastEditedAt: 2026-09-25T05:12:28Z
closedAt: 2022-06-14T19:55:51Z
---

# deps: up go-github to v45.1.0

Update the GitHub client to v45.1.0 while keeping maintainer's user scenarios working. This is routine dependency maintenance: take upstream changes without accumulating a large gap between versions.

The original target release: [v45.1.0](https://github.com/google/go-github/releases/tag/v45.1.0). New client capabilities must not be counted as new maintainer features: they only become available once a command uses them.

The update is done when the project builds, authorization and GitHub data retrieval keep their previous contract, and the checks for the affected scenarios pass. Crossing a major version also requires aligning module and import paths; editing a single version number is not enough.

<!-- 2022-06-14T19:55Z https://github.com/octomation/maintainer/issues/49#issuecomment-1155652368
https://github.com/octomation/maintainer/commit/dd6f6002fe441ec0e417cc62e83d9a853c47563e
-->
