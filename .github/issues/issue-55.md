---
code:
id: I_kwDOE2M9Zc5LnX8m
databaseId: 1268612902
number: 55
url: https://github.com/octomation/maintainer/issues/55
title: "ci/cd: rename coverage to report"
labels: []
milestone:
state: CLOSED
stateReason: COMPLETED
createdAt: 2022-06-12T15:14:08Z
updatedAt: 2022-06-14T19:51:35Z
lastEditedAt:
closedAt: 2022-06-14T19:51:35Z
---

# ci/cd: rename coverage to report

Rename the coverage-upload job from `coverage` to `report` so that the CI stage names consistently denote an action — every job is a verb. That makes the workflow easier to read and separates running the tests from publishing their report.

The expected scheme:

```text
lint   → quality checks
test   → tests and coverage collection
report → sending the report
```

The rename must preserve the job dependencies, the handover of the coverage artifact, and the purpose of the stage; the report must not be lost because a name changed.

**Current state:** the historical task is closed. [ci.yml](../workflows/ci.yml) contains `report` with the display name `Reporting`; it depends on `test` and consumes `code-coverage-report`. No separate user-facing command is required.

<!-- 2022-06-14T19:51Z https://github.com/octomation/maintainer/issues/55#issuecomment-1155648481
https://github.com/octomation/maintainer/commit/af1de01a4b5408e679ba5216b8e4f4bfbc381276
-->
