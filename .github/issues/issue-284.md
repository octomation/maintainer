---
code:
id: I_kwDOE2M9Zc6lESTy
databaseId: 2769364210
number: 284
url: https://github.com/octomation/maintainer/issues/284
title: "github: contribution: incorrect week number"
labels:
  - "scope: code"
  - "type: bug"
  - "severity: minor"
  - "impact: low"
  - "effort: easy"
milestone:
state: OPEN
stateReason:
createdAt: 2025-01-05T17:04:56Z
updatedAt: 2026-09-25T05:11:33Z
lastEditedAt: 2026-09-25T05:11:33Z
closedAt:
---

# github: contribution: incorrect week number

Fix the week number at a year boundary. In lookup, the week beginning 29 December 2024 and containing the start of January 2025 is labelled `#53`, although the expected calendar labels it `#1`.

A reproducible request for the period:

```bash
maintainer github contribution lookup 2024-12-29/1
# Expected header: #01 instead of #53.
```

The original evidence is this [screenshot](https://github.com/user-attachments/assets/66a4020b-8ee7-456e-a5d9-e4a56621b35f).

Consistent numbering of the Sunday columns at the December/January boundary is expected, with the data and the right-hand dates staying in the same cells. An ordinary year transition, a year with 53 ISO weeks, and the detailed suggest that uses the same table all need checking. The defect is specifically in the week label; loading several years is [[issue-43]].
