---
code:
id: I_kwDOE2M9Zc5h0Mjl
databaseId: 1641072869
number: 120
url: https://github.com/octomation/maintainer/issues/120
title: "ci/cd: fix broken build"
labels:
  - "scope: code"
  - "type: bug"
  - "severity: major"
  - "impact: high"
  - "effort: easy"
milestone: "[[milestone-1]]"
state: CLOSED
stateReason: COMPLETED
createdAt: 2023-03-26T20:26:52Z
updatedAt: 2026-09-25T05:09:48Z
lastEditedAt: 2026-09-25T05:09:48Z
closedAt: 2023-03-27T14:37:14Z
---

# ci/cd: fix broken build

Restore a passing project build after CI broke. While the basic check does not work, the maintainer cannot tell a defect in a new change from a malfunction of the tooling and environment.

The original report linked [job 7971632672](https://github.com/octomation/maintainer/actions/runs/4526452170/jobs/7971632672) with no error text. A single link is not enough to attribute the failure to a particular Go version or GitHub action.

**Boundaries:** the scope covers both the code and the `tools/` toolset, including its dependencies and generation (commit `1710249`), not only the workflow syntax error of [[issue-54]].

The expected result is a reproducible build and a passing affected check with consistent dependencies, without masking the failure by disabling a check that is needed.

Related: later infrastructure divergences [[issue-175]].

<!-- 2023-03-26T20:53Z https://github.com/octomation/maintainer/issues/120#issuecomment-1484220094
https://github.com/octomation/maintainer/actions/runs/4526542996/jobs/7971783088#step:6:68
-->
