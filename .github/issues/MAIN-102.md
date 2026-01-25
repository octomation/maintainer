---
code: MAIN-102
id: I_kwDOE2M9Zc5iv7Wc
databaseId: 1656731036
number: 128
url: https://github.com/octomation/maintainer/issues/128
title: "github: contribution: support preset for suggestion"
issueType: null
labels:
  - "scope: code"
  - "scope: test"
  - "type: feature"
  - "effort: medium"
assignees:
  - kamilsk
milestone: "[[MAIN-M1]]"
parent: null
issueFields: {}
state: OPEN
stateReason: null
duplicateOf: null
createdAt: 2023-04-06T06:11:45Z
updatedAt: 2026-09-25T05:10:04Z
lastEditedAt: 2026-09-25T05:10:04Z
closedAt: null
---

# github: contribution: support preset for suggestion

Support a set of contribution target levels instead of a single lower bound. The motivation is a smoother distribution of activity: one unusually high day must not automatically set an arbitrary target for the rest.

The originally proposed set is `5, 10, 15` — in the initial sketch the single basis of the suggestion model became a list of them. A prototype of the interface:

```bash
maintainer github contribution suggest --targets=5,10,15 2021/+10
```

The intended meaning, illustrated: for a week whose maximum is 7, choose the agreed level 10 rather than a new arbitrary level of 7. The exact rule for picking a step, and the behaviour above the last step, have to be defined before implementation; this is an illustration, not an approved algorithm.

The interaction of the single-target `--target` flag with a preset must stay comprehensible, and the chosen target must be shown in the result. The deficit statistics ([[MAIN-79]]) must use the same rules.
