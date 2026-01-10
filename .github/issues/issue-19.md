---
code:
id: MDU6SXNzdWU5ODE1NzYwNzU=
databaseId: 981576075
number: 19
url: https://github.com/octomation/maintainer/issues/19
title: "transition entity"
labels:
  - "scope: code"
milestone:
state: OPEN
stateReason:
createdAt: 2021-08-27T20:30:43Z
updatedAt: 2026-09-25T05:10:56Z
lastEditedAt: 2026-09-25T05:10:56Z
closedAt:
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

Related: comparing configurations [[issue-25]], the change plan in the [fetch specification](<../notes/Specs/GitHub fetcher, PoC implementation plan.md>), label scope removal [[issue-80]].
