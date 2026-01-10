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
updatedAt: 2023-07-15T06:09:08Z
lastEditedAt:
closedAt: 2023-07-15T06:09:07Z
---

# github: contribution: broken heatmap

Restore reading of the calendar after GitHub changed its HTML markup. Once the selector stops finding the cells, every contribution-based command risks receiving an empty map instead of the real activity.

The historical symptom is preserved in the [screenshot from the report](https://github.com/octomation/maintainer/assets/1165416/e3b9917a-0f51-4d27-a232-fa9f5f590563).

**Context established from history:** commit `a9229ce` replaced the `svg.js-calendar-graph-svg rect.ContributionCalendar-day` lookup with `table.ContributionCalendar-grid td.ContributionCalendar-day` and refreshed the stored HTML fixtures. That is a useful localization of the defect, not a requirement to preserve that old format forever.

The expected result: the previous dates and counts are read again from the new response, and a real user's empty calendar cannot be confused with a failure to recognize any elements.

**Current state:** the issue is closed; the current parser uses table cells and additionally links them to a tooltip. The later relocation of the counter is [#174](issue-174.md), and the later change of the loading mechanism is [#220](issue-220.md).
