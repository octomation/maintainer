---
code:
id: I_kwDOE2M9Zc5ien0m
databaseId: 1652194598
number: 123
url: https://github.com/octomation/maintainer/issues/123
title: "github: contribution: suggest and lookup paniced on Sunday"
labels:
  - "scope: code"
  - "type: bug"
  - "severity: critical"
  - "impact: high"
  - "effort: medium"
milestone: "[[milestone-1]]"
state: CLOSED
stateReason: NOT_PLANNED
createdAt: 2023-04-03T14:31:28Z
updatedAt: 2026-09-25T05:09:55Z
lastEditedAt: 2026-09-25T05:09:55Z
closedAt: 2023-04-05T18:50:14Z
---

# github: contribution: suggest and lookup paniced on Sunday

Eliminate the crashes of lookup and suggest on a Sunday while the time window is computed. The user must get either a calendar or a comprehensible input error, whatever the day of the week.

The original conditions: 2 April 2023, 09:09:53, UTC+03:00.

```bash
maintainer github contribution suggest /-20
maintainer github contribution lookup /-20
# recovered: assertion is not a true
# unexpected panic occurred
```

Both stacks pointed at range operations: `ExpandRight` for suggest and `Shift` for lookup. For a bug report that is more useful than the full Cobra/runtime stack.

A correct, non-empty window with Sunday at the start of the week is expected, and clamping the future must not produce an impossible range. Verification has to pin the time and the Git anchor.

Related: another range panic [[issue-155]], a future HEAD [[issue-148]].

<!-- 2023-04-05T18:50Z https://github.com/octomation/maintainer/issues/123#issuecomment-1497964858
I hope the problem has gone. I will check it on Sunday. `maintainer github contribution suggest 2023-04-02/-20` works well.
-->
