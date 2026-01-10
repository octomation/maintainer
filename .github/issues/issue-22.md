---
code:
id: I_kwDOE2M9Zc4-calF
databaseId: 1047636293
number: 22
url: https://github.com/octomation/maintainer/issues/22
title: "update github client"
labels:
  - "scope: code"
milestone:
state: CLOSED
stateReason: COMPLETED
createdAt: 2021-11-08T16:12:16Z
updatedAt: 2026-09-25T05:11:13Z
lastEditedAt: 2026-09-25T05:11:13Z
closedAt: 2022-04-30T19:27:26Z
---

# update github client

Update the GitHub client to v44.0.0 while keeping maintainer's user scenarios working. This is routine dependency maintenance: take upstream changes without accumulating a large gap between versions.

Review the releases in order: [v40.0.0](https://github.com/google/go-github/releases/tag/v40.0.0), [v41.0.0](https://github.com/google/go-github/releases/tag/v41.0.0), [v42.0.0](https://github.com/google/go-github/releases/tag/v42.0.0), [v43.0.0](https://github.com/google/go-github/releases/tag/v43.0.0), [v44.0.0](https://github.com/google/go-github/releases/tag/v44.0.0). New client capabilities must not be counted as new maintainer features: they only become available once a command uses them.

The update is done when the project builds, authorization and GitHub data retrieval keep their previous contract, and the checks for the affected scenarios pass. Crossing a major version also requires aligning module and import paths; editing a single version number is not enough.
