---
code:
id: MDU6SXNzdWU4MjEzMjc2Mzc=
databaseId: 821327637
number: 8
url: https://github.com/octomation/maintainer/issues/8
title: "big picture: labels"
labels:
  - "help wanted"
milestone:
state: CLOSED
stateReason: COMPLETED
createdAt: 2021-03-03T17:31:12Z
updatedAt: 2023-03-31T15:34:19Z
lastEditedAt: 2021-03-15T19:53:00Z
closedAt: 2021-03-15T19:53:10Z
---

# big picture: labels

Reduce work with GitHub labels to a manageable set of rules: read the current labels, choose a target set, and prepare reviewable changes. This removes the need to configure identical categories by hand across many repositories.

Historical scope: the default set (`defaults`) and the target set (`target`) were marked done; a separate mapping step was dropped from the plan.

An example of the result at the user level:

```text
bug        → type: bug
feature    → type: feature
help wanted: keep
```

Related parts: inventory [#12](issue-12.md), presets [#13](issue-13.md), preview [#14](issue-14.md), issue classification [#32](issue-32.md).

**Current state:** the direction is historical and closed. In [#80](issue-80.md) label management was handed over to the settings app; there is no `github labels` group in the current CLI. This issue does not propose bringing it back without a new decision about product boundaries.

<!-- 2021-03-09T18:59Z https://github.com/octomation/maintainer/issues/8#issuecomment-794309846
mapping will be done in #12
-->

<!-- 2021-03-15T19:53Z https://github.com/octomation/maintainer/issues/8#issuecomment-799707784
done by https://miro.com/app/board/o9J_lVCU5K4=/?moveToWidget=3074457355397827885&cot=14
-->
