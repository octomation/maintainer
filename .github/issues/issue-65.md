---
code:
id: I_kwDOE2M9Zc5NIZ29
databaseId: 1294048701
number: 65
url: https://github.com/octomation/maintainer/issues/65
title: "github: contribution: lookup has problem with timezone"
labels:
  - "scope: code"
  - "scope: test"
milestone: "[[milestone-1]]"
state: CLOSED
stateReason: COMPLETED
createdAt: 2022-07-05T09:38:50Z
updatedAt: 2026-09-25T05:13:12Z
lastEditedAt: 2026-09-25T05:13:12Z
closedAt: 2022-07-05T10:34:15Z
---

# github: contribution: lookup has problem with timezone

Reconcile GitHub's calendar dates with the local time zone in lookup. In the original report the terminal and the browser showed a different boundary for the available activity, which made the most recent contributions hard to assess.

The historical reproduction:

```bash
maintainer github contribution lookup /-5
# In the table the available data ends at 2022-07-05.
```

The [GitHub screenshot](https://user-images.githubusercontent.com/1165416/177298172-c6fb76df-e451-450f-ac27-97124eeca477.png) is kept for comparison. Counts are expected to match for one calendar date, and the future must be clearly distinct from zero activity.

Verifying a defect like this requires pinning the moment of the run and the time zone, especially near midnight.

Related: the neighbouring Sunday defects [[issue-66]], configuring working hours [[issue-127]].

<!-- 2022-07-05T10:34Z https://github.com/octomation/maintainer/issues/65#issuecomment-1174903261
it's related to timezone, but I don't have possibility to change it without login on it

<img width="241" alt="image" src="https://user-images.githubusercontent.com/1165416/177308878-8fe7e1b2-0fa0-4230-93c2-f5e8a6f685ed.png">

the solution is to check in private mode
-->
