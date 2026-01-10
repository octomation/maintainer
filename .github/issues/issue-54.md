---
code:
id: I_kwDOE2M9Zc5LnWW1
databaseId: 1268606389
number: 54
url: https://github.com/octomation/maintainer/issues/54
title: "ci/cd: invalid workflow file: .github/workflows/cd.yml#L38"
labels: []
milestone:
state: CLOSED
stateReason: COMPLETED
createdAt: 2022-06-12T14:51:18Z
updatedAt: 2026-09-25T05:12:43Z
lastEditedAt: 2026-09-25T05:12:43Z
closedAt: 2022-06-12T14:57:55Z
---

# ci/cd: invalid workflow file: .github/workflows/cd.yml#L38

Fix the invalid job dependency in the publishing workflow: GitHub rejected the file before the build could run. The user-visible effect is that a release cannot be delivered even when the code itself builds.

The original failure: [run 2483849692](https://github.com/octomation/maintainer/actions/runs/2483849692), pointing at `.github/workflows/cd.yml#L38`.

The expected result is a valid job graph: publish first, then notify based on its outcome.

<!-- 2022-06-12T14:57Z https://github.com/octomation/maintainer/issues/54#issuecomment-1153196709
https://github.com/octomation/maintainer/commit/585ab3676657091cd8122ccdc473c81b683d32f0
-->
