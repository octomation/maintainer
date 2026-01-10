---
code:
id: I_kwDOE2M9Zc6EHV6X
databaseId: 2216517271
number: 219
url: https://github.com/octomation/maintainer/issues/219
title: "ci/cd: Continuous integration healthcheck doesn't work properly"
labels:
  - "scope: test"
  - "type: bug"
  - "severity: critical"
  - "scope: inventory"
  - "impact: high"
  - "effort: easy"
milestone: "[[milestone-1]]"
state: CLOSED
stateReason: COMPLETED
createdAt: 2024-03-30T15:07:52Z
updatedAt: 2026-09-25T05:11:11Z
lastEditedAt: 2026-09-25T05:11:11Z
closedAt: 2024-03-30T17:14:09Z
---

# ci/cd: Continuous integration healthcheck doesn't work properly

Make sure the daily GitHub calendar check actually runs against fresh HTML. The original healthcheck updated the testdata, but the changes never reached the next step, so a green result on stale fixtures could hide a broken integration.

The expected sequence:

```text
Fetch fresh GitHub HTML
→ store the fixtures in the working checkout
→ run the tests against exactly those files
```

A download failure must stop the check rather than quietly fall back to a run against old data. The point of the healthcheck is to notice an external format change that ordinary tests over stored fixtures cannot detect.

Related: [[issue-174]] (the check was introduced after it), [[issue-220]] (the change of how the data is obtained).

<!-- 2024-03-30T17:14Z https://github.com/octomation/maintainer/issues/219#issuecomment-2028312166
fixed by 0c9611551551b638d6d2be5475eddccc93d41b39
-->
