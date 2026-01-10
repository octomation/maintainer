---
code:
id: I_kwDOE2M9Zc5ro_z1
databaseId: 1805909237
number: 150
url: https://github.com/octomation/maintainer/issues/150
title: "github: contribution: broken heatmap"
labels:
  - "scope: code"
  - "type: bug"
  - "severity: critical"
  - "impact: high"
  - "effort: easy"
milestone: "[[milestone-1]]"
state: CLOSED
stateReason: COMPLETED
createdAt: 2023-07-15T05:30:17Z
updatedAt: 2026-09-25T05:10:33Z
lastEditedAt: 2026-09-25T05:10:33Z
closedAt: 2023-07-15T06:09:07Z
---

# github: contribution: broken heatmap

Restore reading of the calendar after GitHub changed its HTML markup. Once the selector stops finding the cells, every contribution-based command risks receiving an empty map instead of the real activity.

The historical symptom is preserved in the [screenshot from the report](https://github.com/octomation/maintainer/assets/1165416/e3b9917a-0f51-4d27-a232-fa9f5f590563).

The expected result: the previous dates and counts are read again from the new response, and a real user's empty calendar cannot be confused with a failure to recognize any elements.

Source crumb: commit `a9229ce` replaced the `svg.js-calendar-graph-svg rect.ContributionCalendar-day` lookup with `table.ContributionCalendar-grid td.ContributionCalendar-day` and refreshed the stored HTML fixtures. That localizes the defect; it is not a requirement to preserve that format.

Related: [[issue-174]] (the later relocation of the counter), [[issue-220]] (the later change of the loading mechanism).
