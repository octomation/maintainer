---
id: 19
database_id: 981576075
node_id: MDU6SXNzdWU5ODE1NzYwNzU=
status: open
title: "transition entity"
labels: ["scope: code"]
url: https://github.com/octomation/maintainer/issues/19
created_at: 2021-08-27T20:30:43Z
updated_at: 2023-03-31T15:34:35Z
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

**Current state:** there is no separate transition model, and labels were removed ([#80](issue-80.md)). The idea is useful for comparing configurations ([#25](issue-25.md)) and agrees with the change plan in the [fetch specification](<../notes/Specs/GitHub fetcher, PoC implementation plan.md>). This task defines the general meaning of a change plan; it does not require bringing labels back or introducing a particular internal structure.
