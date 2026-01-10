---
code:
id: I_kwDOE2M9Zc57frK6
databaseId: 2071900858
number: 193
url: https://github.com/octomation/maintainer/issues/193
title: "github: suggestion: reduce jitter"
labels:
  - "scope: code"
  - "type: improvement"
  - "impact: high"
  - "effort: easy"
milestone: "[[milestone-1]]"
state: OPEN
stateReason:
createdAt: 2024-01-09T08:49:32Z
updatedAt: 2024-01-09T08:49:33Z
lastEditedAt:
closedAt:
---

# github: suggestion: reduce jitter

Reduce the random offset of the suggestion time and tie it to the target activity. A large offset after every commit quickly consumes the day's available working time, especially with a high target.

The scenario, illustrated: with a target of 50 contributions, an hour is left until the end of the working interval; yet another nearly hour-long offset leaves too little time for the remaining work.

The proposal is to choose the offset with the target in mind and to cap it from above. How the remaining contributions and the time until the boundary are taken into account has to be defined, as does what happens when no room is left. The resulting date must not go beyond the agreed working interval or the present moment.

**At present** [suggest](../../internal/command/github/contribution/suggest.go) adds a random value of up to an hour after the time has already been chosen, with no re-check of the upper bound. The original point of discussion is the [historical jitter call](https://github.com/octomation/maintainer/blob/020d048eebf8059814ca342b4ccd1e34a87784c1/internal/command/github/contribution/suggest.go#L50). Related: configuring hours [#127](issue-127.md), a future HEAD [#133](issue-133.md).
