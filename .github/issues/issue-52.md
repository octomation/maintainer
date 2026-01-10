---
code:
id: I_kwDOE2M9Zc5Ll5En
databaseId: 1268224295
number: 52
url: https://github.com/octomation/maintainer/issues/52
title: "tools: experimental: integrate slides and describe v0.1.x using it"
labels:
  - "scope: docs"
  - "type: feature"
  - "scope: deps"
  - "scope: inventory"
  - "impact: medium"
  - "effort: medium"
milestone: "[[milestone-1]]"
state: OPEN
stateReason:
createdAt: 2022-06-11T09:59:51Z
updatedAt: 2026-09-25T05:12:38Z
lastEditedAt: 2026-09-25T05:12:37Z
closedAt:
---

# tools: experimental: integrate slides and describe v0.1.x using it

Try telling the story of a release through a sequence of user stories in terminal slides. A plain list of commands does not show how they solve a task end to end — the motivation is a storytelling-based changelog.

The experiment: describe the v0.1.x series with [slides](https://github.com/maaslalani/slides). The proposed location for the source is `docs/changelog/v0.1.x`.

A prototype of the storyline:

```text
1. Find the gaps in the calendar: lookup.
2. Choose a date to contribute: suggest.
3. Store the starting point: snapshot.
4. Show the result: diff.
```

The experiment is ready when the presentation can be reproduced and its examples match the real CLI. Whether the format actually helps to understand the scenario should be assessed before extending it to the remaining releases. The experiment does not replace the concise user reference under `docs/`.

<!-- 2023-04-01T14:07Z https://github.com/octomation/maintainer/issues/52#issuecomment-1492980304
Add `howto` command.
-->
