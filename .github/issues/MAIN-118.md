---
code: MAIN-118
id: I_kwDOE2M9Zc57frK6
databaseId: 2071900858
number: 193
url: https://github.com/octomation/maintainer/issues/193
title: "github: suggestion: reduce jitter"
issueType: null
labels:
  - "scope: code"
  - "type: improvement"
  - "impact: high"
  - "effort: easy"
assignees:
  - kamilsk
milestone: "[[MAIN-M1]]"
parent: null
issueFields: {}
state: OPEN
stateReason: null
duplicateOf: null
createdAt: 2024-01-09T08:49:32Z
updatedAt: 2026-11-27T14:37:58Z
lastEditedAt: 2026-11-27T14:37:58Z
closedAt: null
---

# github: suggestion: reduce jitter

Reduce the random offset of the suggestion time and tie it to the target activity. A large offset after every commit quickly consumes the day's available working time, especially with a high target.

The scenario, illustrated: with a target of 50 contributions, an hour is left until the end of the working interval; yet another nearly hour-long offset leaves too little time for the remaining work.

The proposal is to choose the offset with the target in mind and to cap it from above. How the remaining contributions and the time until the boundary are taken into account has to be defined, as does what happens when no room is left. The resulting date must not go beyond the agreed working interval or the present moment.

The original point of discussion is the [historical jitter call](https://github.com/octomation/maintainer/blob/020d048eebf8059814ca342b4ccd1e34a87784c1/internal/command/github/contribution/suggest.go#L50). Related: configuring hours [[MAIN-101]], a future HEAD [[MAIN-106]].

## Acceptance criteria

- [ ] The offset is at most `min(1 hour, available / remaining)`, where
      `remaining` is the effective target minus the actual count of the day
      and `available` is the time from the anchor to the earlier of now and
      the end of the working interval.
- [ ] With 50 contributions remaining and one hour available, the offset is
      under 72 seconds; with no time available, there is no offset.
- [ ] Tests cover 1, 5 and 50 remaining, a very large target, the bounds of
      the interval and rounding to whole seconds; the suggestion never leaves
      the interval or passes now, and stdout stays one timestamp.
- [ ] `docs/content/contributions.md` says how the offset is chosen.
