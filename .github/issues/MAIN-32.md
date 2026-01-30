---
code: MAIN-32
id: I_kwDOE2M9Zc5JeZL9
databaseId: 1232704253
number: 32
url: https://github.com/octomation/maintainer/issues/32
title: "command: labels classify issue"
issueType: null
labels: []
assignees:
  - kamilsk
milestone: "[[MAIN-M1]]"
parent: null
issueFields: {}
state: CLOSED
stateReason: NOT_PLANNED
duplicateOf: null
createdAt: 2022-05-11T14:08:29Z
updatedAt: 2026-10-20T07:32:15Z
lastEditedAt: 2026-10-19T07:22:38Z
closedAt: 2023-03-31T15:32:01Z
---

# command: labels classify issue

Help the author of an issue choose consistent labels through a short step-by-step form. This lowers the barrier to the agreed classification and makes tasks comparable during prioritization.

A historical prototype of the interaction:

```text
Is it a bug?              Yes
What does it affect?      Code
How large is the impact?  Medium
Proposal: type: bug, scope: code, impact: medium
```

The form must explain what the categories mean, show the resulting set, and allow the answers to be corrected before anything is applied. Task type, scope, impact and effort must stay distinguishable rather than be assigned automatically from a single answer.

Related: [[MAIN-80]].
