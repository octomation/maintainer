---
code:
id: I_kwDOE2M9Zc52_zI_
databaseId: 1996436031
number: 174
url: https://github.com/octomation/maintainer/issues/174
title: "github: contribution: html markup was changed"
labels:
  - "scope: code"
  - "scope: test"
  - "type: bug"
  - "severity: critical"
  - "scope: inventory"
  - "impact: high"
  - "effort: medium"
milestone: "[[milestone-1]]"
state: CLOSED
stateReason: COMPLETED
createdAt: 2023-11-16T09:29:52Z
updatedAt: 2026-09-25T05:10:47Z
lastEditedAt: 2026-09-25T05:10:47Z
closedAt: 2023-11-16T10:37:52Z
---

# github: contribution: html markup was changed

Restore the contribution counts after GitHub moved their text into separate tooltip elements. The dates kept resolving, so the defect looked like a calendar without activity rather than an outright load failure.

The original reproduction:

```bash
maintainer github contribution lookup 2022-09-18/3
# Every count is rendered as "-".
```

The useful detail of the format: the date lives in `<td data-date="…" id="…">`, while the number of contributions lives in `<tool-tip for="…">`. They have to be joined by identifier; the tooltip cannot be assumed to sit next to its cell.

Exact restoration of the dates and numbers is expected, days without activity included. Besides fixing the parsing, the original checklist covered a daily integration check and the removal of legacy markers for previous formats.

Related: [[issue-219]] (passing fresh fixtures further down CI).
