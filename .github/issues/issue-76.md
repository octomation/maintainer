---
code:
id: I_kwDOE2M9Zc5O0LuO
databaseId: 1322302350
number: 76
url: https://github.com/octomation/maintainer/issues/76
title: "github: contribution: support strict ISO 8601 format as input date"
labels:
  - "scope: code"
  - "type: feature"
  - "impact: medium"
  - "effort: easy"
milestone: "[[milestone-1]]"
state: CLOSED
stateReason: COMPLETED
createdAt: 2022-07-29T14:10:53Z
updatedAt: 2026-09-25T05:13:40Z
lastEditedAt: 2026-09-25T05:13:40Z
closedAt: 2023-03-26T20:22:54Z
---

# github: contribution: support strict ISO 8601 format as input date

Accept the full author date of a commit, time zone included, so that suggest takes the time within the day into account. A bare date is not enough: a new commit must not accidentally land before an existing one merely because hours and minutes were lost.

The working example:

```bash
maintainer github contribution suggest --short "$(git --no-pager log -1 --format='%aI')"
```

An RFC 3339 timestamp is expected to be parsed, and the time chosen with respect to the reference moment and the working interval. For a new day, any time inside the working range fits; for the same day, an acceptable offset after the previous commit matters.

Related: configuring working hours and the zone [[issue-127]], the size of the random offset [[issue-193]].
