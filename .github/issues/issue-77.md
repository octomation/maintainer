---
code:
id: I_kwDOE2M9Zc5PUv2Y
databaseId: 1330838936
number: 77
url: https://github.com/octomation/maintainer/issues/77
title: "github: contribution: snapshot and diff must support multiple date input"
labels:
  - "scope: docs"
  - "scope: code"
  - "scope: test"
  - "type: feature"
  - "impact: medium"
  - "effort: easy"
milestone: "[[milestone-1]]"
state: OPEN
stateReason:
createdAt: 2022-08-06T20:07:58Z
updatedAt: 2026-09-25T05:13:43Z
lastEditedAt: 2026-09-25T05:13:43Z
closedAt:
---

# github: contribution: snapshot and diff must support multiple date input

Store contributions for several years in a single call, and compare such combined snapshots. This is needed for daily accounting across a year boundary: changes can touch both the current and the previous year.

The proposed interface:

```bash
maintainer github contribution snapshot 2021 2022 > snapshot.json
maintainer github contribution diff before.json after.json
```

One document holding every requested date is expected, with no repetitions; a year given twice must not duplicate the data. A failure to load one period must be visible rather than hidden behind the appearance of a complete snapshot. The syntax for comparing a file against several live periods still has to be defined.

The main consumer is [[issue-78]] — it simplifies the cronjob recipe described there.

Related: merging years inside the service range [[issue-43]].
