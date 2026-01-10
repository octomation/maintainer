---
code:
id: I_kwDOE2M9Zc5JpSY5
databaseId: 1235559993
number: 37
url: https://github.com/octomation/maintainer/issues/37
title: "deps: up go-github to v44.1.0"
labels:
  - "scope: code"
milestone:
state: CLOSED
stateReason: COMPLETED
createdAt: 2022-05-13T18:31:22Z
updatedAt: 2026-09-25T05:11:57Z
lastEditedAt: 2026-09-25T05:11:57Z
closedAt: 2022-05-13T19:12:20Z
---

# deps: up go-github to v44.1.0

Update the GitHub client to v44.1.0 while keeping maintainer's user scenarios working. This is routine dependency maintenance: take upstream changes without accumulating a large gap between versions.

The original target release: [v44.1.0](https://github.com/google/go-github/releases/tag/v44.1.0). New client capabilities must not be counted as new maintainer features: they only become available once a command uses them.

The update is done when the project builds, authorization and GitHub data retrieval keep their previous contract, and the checks for the affected scenarios pass. Crossing a major version also requires aligning module and import paths; editing a single version number is not enough.

<!-- 2022-05-13T19:12Z https://github.com/octomation/maintainer/issues/37#issuecomment-1126365541
https://github.com/octomation/maintainer/commit/08ca2968337aa8341a0459d0a6f82752b7d5ee4d
-->
