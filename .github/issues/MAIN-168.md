---
code: MAIN-168
id: I_kwDOE2M9Zc8AAAABi07ocA
databaseId: 6632171632
number: 570
url: https://github.com/octomation/maintainer/issues/570
title: "github: contribution: fetch again a month cached while it was open"
issueType: Bug
labels:
  - "scope: code"
  - "scope: test"
  - "type: bug"
  - "severity: minor"
  - "impact: medium"
  - "effort: medium"
assignees:
  - kamilsk
milestone: null
parent: null
issueFields: {}
state: CLOSED
stateReason: COMPLETED
duplicateOf: null
createdAt: 2026-11-26T17:41:55Z
updatedAt: 2026-11-26T20:01:38Z
lastEditedAt: 2026-11-26T20:01:38Z
closedAt: 2026-11-26T19:43:12Z
---

# github: contribution: fetch again a month cached while it was open

## Problem

On the last day of a month, the event lists of GitHub lag behind the total, so a
run caches the month with the days of the API calendar. After the month closes,
a run with an equal total reuses that calendar for good. A review near midnight
then keeps its day in the zone of the account, and no warning names the month.

## PoC

1. On 2026-10-31 at 23:00Z a run caches October with one contribution that has
   no date yet, so October takes the days of the API calendar.
2. On 2026-11-01 the same contribution is a review on 2026-10-20, and the total
   of October does not change.
3. Today: October keeps the API calendar, and no month is fetched. Expected:
   October is fetched again, and the review counts on 2026-10-20.

## Expected behavior

A calendar taken while its month was open is not reused after the month closes.
The first run after the close fetches that month again. Later runs reuse that
result as for any closed month.

## Acceptance criteria

- [x] A test of the PoC steps fetches October on 2026-11-01 and counts the
      review on 2026-10-20.
- [x] A warm run over closed months that were cached after their close fetches
      no month.

## Out of scope

Months with contributions that still have no date after the close: they keep the
rule of [[MAIN-159]]. The edges of the monthly slices: [[MAIN-160]].

Source: [20261125T152232Z](../reports/20261125T152232Z.md), "A month cached with the API calendar while open keeps it after it closes".
