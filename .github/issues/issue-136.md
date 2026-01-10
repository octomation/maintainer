---
code:
id: I_kwDOE2M9Zc5j5gT9
databaseId: 1676018941
number: 136
url: https://github.com/octomation/maintainer/issues/136
title: "github: contribution: incorrect suggest for today"
labels:
  - "scope: code"
  - "type: bug"
  - "severity: critical"
  - "impact: high"
  - "effort: easy"
milestone: "[[milestone-1]]"
state: OPEN
stateReason:
createdAt: 2023-04-20T05:19:35Z
updatedAt: 2023-04-20T09:18:36Z
lastEditedAt:
closedAt:
---

# github: contribution: incorrect suggest for today

Investigate the incorrect suggestion for the current day in combination with a Git wrapper. In the original report the detailed message lost the date, and the calendar was hard to reconcile with the commit that followed.

The observation of 20 April 2023:

```text
git contrib docs: readme: improve headline
Suggestion is , 0 → 10
Created commit:  2023-04-20 08:29:24 +0300
Previous commit: 2023-04-19 23:25:37 +0300
```

The calendar showed 8 for 19 April and a highlighted empty cell for 20 April. The difference between the days does not by itself prove a counter defect: the chosen date, the time zone and the moment the data was fetched all have to be checked.

Consistency between the timestamp, the highlighted day and actual/target is expected, within the time boundaries. **Current code context:** the date goes to stdout and the explanation to stderr, so under shell substitution the date predictably disappears from the visible message. A direct `suggest --short git/+1` call and the external wrapper have to be checked separately, rather than attributing the lost text to the algorithm without verification. Future-HEAD cases are [#133](issue-133.md).

<!-- 2023-04-20T09:18Z https://github.com/octomation/maintainer/issues/136#issuecomment-1515999107
the key is 23:25:37 -> outside working hours.
-->
