---
code:
id: I_kwDOE2M9Zc5MRRyu
databaseId: 1279597742
number: 62
url: https://github.com/octomation/maintainer/issues/62
title: "deps: up github.com/google/go-github to v45.2.0"
labels:
  - "scope: code"
milestone:
state: CLOSED
stateReason: COMPLETED
createdAt: 2022-06-22T06:06:06Z
updatedAt: 2022-06-22T06:07:52Z
lastEditedAt:
closedAt: 2022-06-22T06:07:52Z
---

# deps: up github.com/google/go-github to v45.2.0

Update the GitHub client to v45.2.0 while keeping maintainer's user scenarios working. This is routine dependency maintenance: take upstream changes without accumulating a large gap between versions.

The original target release: [v45.2.0](https://github.com/google/go-github/releases/tag/v45.2.0). New client capabilities must not be counted as new maintainer features: they only become available once a command uses them.

The update is done when the project builds, authorization and GitHub data retrieval keep their previous contract, and the checks for the affected scenarios pass. Crossing a major version also requires aligning module and import paths; editing a single version number is not enough.

**Current state:** this historical issue is closed. [go.mod](../../go.mod) already uses `github.com/google/go-github/v91 v91.0.0`, so returning the dependency to v45.2.0 is not required. This is the record of one concrete update step, not a request to downgrade.

<!-- 2022-06-22T06:07Z https://github.com/octomation/maintainer/issues/62#issuecomment-1162683489
https://github.com/octomation/maintainer/commit/0ad159e764a4ad9c15b88fe2ace249bd2b5acbd1
-->
