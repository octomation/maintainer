---
code: MAIN-19
id: MDU6SXNzdWU5ODE1NzYwNzU=
databaseId: 981576075
number: 19
url: https://github.com/octomation/maintainer/issues/19
title: "transition entity"
issueType: null
labels:
  - "scope: code"
assignees:
  - kamilsk
milestone: "[[MAIN-M2]]"
parent: null
issueFields: {}
state: OPEN
stateReason: null
duplicateOf: null
createdAt: 2021-08-27T20:30:43Z
updatedAt: 2026-10-21T10:21:58Z
lastEditedAt: 2026-10-19T07:22:31Z
closedAt: null
---

# transition entity

Describe a change to an entity as a transition from a source state to a target state. That makes a meaningful dry-run possible and allows exactly the agreed changes to be applied, without overwriting fields the task does not touch.

The example from the original discussion, using labels:

```text
LabelX { name: "x", color: "000" }
LabelY { name: "y", color: "000" }

Transition { name: { from: "x", to: "y" } }
# The color is unchanged and is not part of the transition.
```

For the user the plan reads as "rename x to y; keep the remaining properties". If the source state has changed in the meantime, the divergence must be shown rather than the stale plan presented as current.

This task defines the general meaning of a change plan; it does not require bringing labels back or introducing a particular internal structure.

Related: comparing configurations [[MAIN-25]], the [fetch guide](https://maintainer.octolab.org/fetch/), label scope removal [[MAIN-80]].
