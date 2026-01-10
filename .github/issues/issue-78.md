---
code:
id: I_kwDOE2M9Zc5PUv4q
databaseId: 1330839082
number: 78
url: https://github.com/octomation/maintainer/issues/78
title: "github: contribution: tips and tricks, daily snapshot"
labels:
  - "scope: docs"
  - "type: feature"
  - "scope: inventory"
  - "impact: medium"
  - "effort: medium"
milestone: "[[milestone-1]]"
state: OPEN
stateReason:
createdAt: 2022-08-06T20:08:51Z
updatedAt: 2026-09-25T05:13:45Z
lastEditedAt: 2026-09-25T05:13:45Z
closedAt:
---

# github: contribution: tips and tricks, daily snapshot

Document a convenient way to track contributions daily: store a starting point and later see what changed. This answers "what was my contribution impact today", including changes to older calendar dates.

The scenario for a single year:

```bash
maintainer github contribution snapshot 2022 > daily.2022.json
# Later:
maintainer github contribution diff daily.2022.json 2022
```

A scheduled run must keep a successfully fetched snapshot and must not overwrite the previous good file on failure. The path has to be suitable for storage, and the time of the daily run has to be chosen explicitly by the user.

**Proposed follow-up:** a combined snapshot of the previous and the current year depends on [[issue-77]]. The short `diff progress` alias comes from the original idea, and the `cron` and `year` calls in the original example were notional, not part of maintainer. A finished recipe has to rest on a real scheduler and on the available CLI.

<!-- 2023-04-01T14:03Z https://github.com/octomation/maintainer/issues/78#issuecomment-1492978937
do it by GitHub Actions
-->
