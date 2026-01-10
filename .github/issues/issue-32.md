---
code:
id: I_kwDOE2M9Zc5JeZL9
databaseId: 1232704253
number: 32
url: https://github.com/octomation/maintainer/issues/32
title: "command: labels classify issue"
labels: []
milestone:
state: CLOSED
stateReason: NOT_PLANNED
createdAt: 2022-05-11T14:08:29Z
updatedAt: 2023-03-31T15:34:20Z
lastEditedAt:
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

**Current state:** the issue is closed and there is no interactive command; the label scope was removed ([#80](issue-80.md)). The motivation for end-to-end prioritization survives in [delayed automation](<../notes/Issues/delayed automation.md>) and [delayed draft](<../notes/Issues/delayed draft.md>), but that is separate future work, not an already implemented continuation of the form.
