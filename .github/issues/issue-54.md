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
updatedAt: 2022-06-12T14:57:55Z
lastEditedAt:
closedAt: 2022-06-12T14:57:55Z
---

# ci/cd: invalid workflow file: .github/workflows/cd.yml#L38

Fix the invalid job dependency in the publishing workflow: GitHub rejected the file before the build could run. The user-visible effect is that a release cannot be delivered even when the code itself builds.

The original failure: [run 2483849692](https://github.com/octomation/maintainer/actions/runs/2483849692), pointing at `.github/workflows/cd.yml#L38`.

**Context established from history:** commit `585ab36` replaced the notification job's dependency on a non-existent `test` job with the existing publishing job. The expected result is a valid job graph: publish first, then notify based on its outcome.

**Current state:** the issue is closed. The job is now called `release`, and `notify` depends on `release` in [cd.yml](../workflows/cd.yml). That confirms the dependency was fixed, but it is not a check that today's release delivery succeeds.

<!-- 2022-06-12T14:57Z https://github.com/octomation/maintainer/issues/54#issuecomment-1153196709
https://github.com/octomation/maintainer/commit/585ab3676657091cd8122ccdc473c81b683d32f0
-->
