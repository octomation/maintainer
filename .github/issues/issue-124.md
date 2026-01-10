---
code:
id: I_kwDOE2M9Zc5igEVn
databaseId: 1652573543
number: 124
url: https://github.com/octomation/maintainer/issues/124
title: "github: contribution: suggest use incorrect center"
labels:
  - "type: bug"
  - "severity: major"
  - "impact: medium"
  - "effort: medium"
milestone: "[[milestone-1]]"
state: CLOSED
stateReason: COMPLETED
createdAt: 2023-04-03T18:36:14Z
updatedAt: 2023-04-05T19:40:33Z
lastEditedAt:
closedAt: 2023-04-05T19:40:32Z
---

# github: contribution: suggest use incorrect center

Centre the detailed suggest calendar on the chosen date. When the highlighted day is off centre, the user does not get the surroundings they expected for checking the suggestion visually.

In the original report the window spanned 7 August – 22 October 2022, columns #32–#42. The suggestion referred to week #38, yet the visual centre turned out to be #37.

Symmetric surroundings are expected for the centred mode, with the reference week positioned correctly, Sunday included. If the window is clamped by the current time, the shortening must be explained by the data boundary rather than by a hidden one-week shift.

**Current state:** the issue is closed; the [LookupRange test](../../internal/model/github/contribution/helper_test.go) contains `issue#124: correct centering`. Directed windows have a different contract ([#41](issue-41.md)); highlighting the chosen day is [#63](issue-63.md).
