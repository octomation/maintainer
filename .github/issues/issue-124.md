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
updatedAt: 2026-09-25T05:09:57Z
lastEditedAt: 2026-09-25T05:09:57Z
closedAt: 2023-04-05T19:40:32Z
---

# github: contribution: suggest use incorrect center

Centre the detailed suggest calendar on the chosen date. When the highlighted day is off centre, the user does not get the surroundings they expected for checking the suggestion visually.

In the original report the window spanned 7 August – 22 October 2022, columns [[issue-32]]–[[issue-42]]. The suggestion referred to week [[issue-38]], yet the visual centre turned out to be [[issue-37]].

Symmetric surroundings are expected for the centred mode, with the reference week positioned correctly, Sunday included. If the window is clamped by the current time, the shortening must be explained by the data boundary rather than by a hidden one-week shift.

Related: directed windows [[issue-41]], highlighting the chosen day [[issue-63]].
