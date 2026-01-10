---
code:
id: I_kwDOE2M9Zc5cTZk7
databaseId: 1548589371
number: 99
url: https://github.com/octomation/maintainer/issues/99
title: "github: contribution: invalid suggestion for 2022 year"
labels:
  - "type: bug"
  - "severity: critical"
  - "impact: high"
  - "effort: medium"
milestone: "[[milestone-1]]"
state: CLOSED
stateReason: NOT_PLANNED
createdAt: 2023-01-19T06:32:01Z
updatedAt: 2026-09-25T05:14:29Z
lastEditedAt: 2026-09-25T05:14:29Z
closedAt: 2023-01-19T18:15:21Z
---

# github: contribution: invalid suggestion for 2022 year

Fix the choice of day when searching over a year: on the data of the original report, suggest skipped over an earlier suitable gap in the calendar.

The recorded divergence:

```text
Obtained: 2022-02-24
Expected: 2022-01-30
```

The original evidence: [the wrong suggestion](https://user-images.githubusercontent.com/1165416/213371611-1adfce14-0ad1-4970-bd83-d3137a981367.png), [the expected day](https://user-images.githubusercontent.com/1165416/213371747-a2574264-8b07-45aa-83c6-06e1ef28bc84.png).

The expected result is a sequential search from the chosen lower bound, respecting the available activity and the target, with no unexplained skipping of a suitable day. An exact check needs the calendar snapshot, the arguments and the moment of the original run; a single screenshot is not a complete reproducible data set.

Related: zero days [[issue-68]], a choice before the reference date [[issue-119]].

<!-- 2023-01-19T18:15Z https://github.com/octomation/maintainer/issues/99#issuecomment-1397411152
isn't a bug

<img width="1392" alt="image" src="https://user-images.githubusercontent.com/1165416/213526844-9815c0d0-4d5a-4d2c-abfd-7424070c89ed.png">

```html
        <g transform="translate(70, 0)">
            <rect width="10" height="10" x="9" y="0" class="ContributionCalendar-day" data-date="2022-01-30" data-level="3" rx="2" ry="2">9 contributions on January 30, 2022</rect>
            <rect width="10" height="10" x="9" y="13" class="ContributionCalendar-day" data-date="2022-01-31" data-level="3" rx="2" ry="2">9 contributions on January 31, 2022</rect>
            <rect width="10" height="10" x="9" y="26" class="ContributionCalendar-day" data-date="2022-02-01" data-level="3" rx="2" ry="2">9 contributions on February 1, 2022</rect>
            <rect width="10" height="10" x="9" y="39" class="ContributionCalendar-day" data-date="2022-02-02" data-level="3" rx="2" ry="2">9 contributions on February 2, 2022</rect>
            <rect width="10" height="10" x="9" y="52" class="ContributionCalendar-day" data-date="2022-02-03" data-level="3" rx="2" ry="2">9 contributions on February 3, 2022</rect>
            <rect width="10" height="10" x="9" y="65" class="ContributionCalendar-day" data-date="2022-02-04" data-level="3" rx="2" ry="2">9 contributions on February 4, 2022</rect>
            <rect width="10" height="10" x="9" y="78" class="ContributionCalendar-day" data-date="2022-02-05" data-level="3" rx="2" ry="2">9 contributions on February 5, 2022</rect>
        </g>
```
-->
