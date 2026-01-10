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
updatedAt: 2026-09-25T05:13:50Z
lastEditedAt: 2026-09-25T05:13:50Z
closedAt: 2021-03-15T19:53:10Z
---

# big picture: labels

Reduce work with GitHub labels to a manageable set of rules: read the current labels, choose a target set, and prepare reviewable changes. This removes the need to configure identical categories by hand across many repositories.

Scope: the default set (`defaults`) and the target set (`target`). A separate mapping step is out of scope.

An example of the result at the user level:

```text
bug        → type: bug
feature    → type: feature
help wanted: keep
```

This issue does not propose bringing label management back without a new decision about product boundaries.

Related: inventory [[issue-12]], presets [[issue-13]], preview [[issue-14]], issue classification [[issue-32]], label scope removal [[issue-80]].

<!-- 2021-03-09T18:59Z https://github.com/octomation/maintainer/issues/8#issuecomment-794309846
mapping will be done in #12
-->

<!-- 2021-03-15T19:53Z https://github.com/octomation/maintainer/issues/8#issuecomment-799707784
done by https://miro.com/app/board/o9J_lVCU5K4=/?moveToWidget=3074457355397827885&cot=14
-->
