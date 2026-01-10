---
code:
id: MDU6SXNzdWU5ODE1NzQxMDE=
databaseId: 981574101
number: 18
url: https://github.com/octomation/maintainer/issues/18
title: "update github client"
labels:
  - "scope: code"
milestone:
state: CLOSED
stateReason: COMPLETED
createdAt: 2021-08-27T20:27:18Z
updatedAt: 2026-09-25T05:10:51Z
lastEditedAt: 2026-09-25T05:10:51Z
closedAt: 2021-10-24T13:04:40Z
---

# update github client

Update the GitHub client to v39.2.0 while keeping maintainer's user scenarios working. This is routine dependency maintenance: take upstream changes without accumulating a large gap between versions.

Review the releases in order: [v37.0.0](https://github.com/google/go-github/releases/tag/v37.0.0), [v38.0.0](https://github.com/google/go-github/releases/tag/v38.0.0), [v38.1.0](https://github.com/google/go-github/releases/tag/v38.1.0), [v39.0.0](https://github.com/google/go-github/releases/tag/v39.0.0), [v39.1.0](https://github.com/google/go-github/releases/tag/v39.1.0), [v39.2.0](https://github.com/google/go-github/releases/tag/v39.2.0). New client capabilities must not be counted as new maintainer features: they only become available once a command uses them.

The update is done when the project builds, authorization and GitHub data retrieval keep their previous contract, and the checks for the affected scenarios pass. Crossing a major version also requires aligning module and import paths; editing a single version number is not enough.
