---
id: 120
database_id: 1641072869
node_id: I_kwDOE2M9Zc5h0Mjl
status: closed
title: "ci/cd: fix broken build"
labels: ["scope: code","type: bug","severity: major","impact: high","effort: easy"]
url: https://github.com/octomation/maintainer/issues/120
created_at: 2023-03-26T20:26:52Z
updated_at: 2023-03-27T14:37:15Z
---

# ci/cd: fix broken build

Restore a passing project build after CI broke. While the basic check does not work, the maintainer cannot tell a defect in a new change from a malfunction of the tooling and environment.

The original report linked [job 7971632672](https://github.com/octomation/maintainer/actions/runs/4526452170/jobs/7971632672) with no error text. A single link is not enough to attribute the failure to a particular Go version or GitHub action.

**Context from history:** commit `1710249` touched both the code and the `tools/` toolset, including its dependencies and generation. The task is therefore broader than the workflow syntax error of [#54](issue-54.md).

The expected result is a reproducible build and a passing affected check with consistent dependencies, without masking the failure by disabling a check that is needed.

**Current state:** the issue is closed. The historical fix does not prove that every current job works; new infrastructure divergences are handled in [#175](issue-175.md).
