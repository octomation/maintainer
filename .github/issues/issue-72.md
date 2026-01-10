---
code:
id: I_kwDOE2M9Zc5Ogo53
databaseId: 1317178999
number: 72
url: https://github.com/octomation/maintainer/issues/72
title: "github: contribution: bad suggestion with Sunday shift"
labels:
  - "scope: code"
  - "scope: test"
milestone: "[[milestone-1]]"
state: CLOSED
stateReason: COMPLETED
createdAt: 2022-07-25T18:01:46Z
updatedAt: 2022-07-25T18:33:26Z
lastEditedAt:
closedAt: 2022-07-25T18:33:26Z
---

# github: contribution: bad suggestion with Sunday shift

Do not shift the suggestion search into the previous week when the reference date is a Sunday. Otherwise a "forward from this date" request proposes a contribution in the past.

The original reproduction:

```bash
maintainer github contribution suggest --delta 2022-05-01/+1
# Obtained: 2022-04-24, 3 → 6
# Expected on the data in the report: 2022-05-04, 1 → 6
```

A request starting from 2 May chose the right day, so the week boundary is exactly what matters. Sunday 1 May is expected to belong to the week that begins, and the chosen day must not precede the reference moment.

**Current state:** the issue is closed; the current calculation handles Sunday as a separate case. Verification needs fixed calendar data and the same period for a Sunday and for a Monday. The window narrowing that followed the attempted fix is described separately in [#73](issue-73.md).
