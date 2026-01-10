---
code:
id: I_kwDOE2M9Zc5iv7Wc
databaseId: 1656731036
number: 128
url: https://github.com/octomation/maintainer/issues/128
title: "github: contribution: support preset for suggestion"
labels:
  - "scope: code"
  - "scope: test"
  - "type: feature"
  - "effort: medium"
milestone: "[[milestone-1]]"
state: OPEN
stateReason:
createdAt: 2023-04-06T06:11:45Z
updatedAt: 2023-04-06T06:11:45Z
lastEditedAt:
closedAt:
---

# github: contribution: support preset for suggestion

Support a set of contribution target levels instead of a single lower bound. The motivation is a smoother distribution of activity: one unusually high day must not automatically set an arbitrary target for the rest.

The originally proposed set is `5, 10, 15` — in the initial sketch the single basis of the suggestion model became a list of them. A prototype of the interface; the flag does not exist yet:

```bash
maintainer github contribution suggest --targets=5,10,15 2021/+10
```

The intended meaning, illustrated: for a week whose maximum is 7, choose the agreed level 10 rather than a new arbitrary level of 7. The exact rule for picking a step, and the behaviour above the last step, have to be defined before implementation; this is an illustration, not an approved algorithm.

**At present** only `--target` is available, defaulting to 5, and the effective target can rise to the week maximum. The interaction of the old flag with a preset must stay comprehensible, and the chosen target must be shown in the result. The deficit statistics ([#79](issue-79.md)) must use the same rules.
