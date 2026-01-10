---
code:
id: I_kwDOE2M9Zc5iv86N
databaseId: 1656737421
number: 129
url: https://github.com/octomation/maintainer/issues/129
title: "github: contribution: calculate contrib to define better day for commits"
labels:
  - "scope: code"
  - "scope: test"
  - "type: feature"
  - "impact: high"
  - "effort: medium"
milestone: "[[milestone-1]]"
state: OPEN
stateReason:
createdAt: 2023-04-06T06:18:30Z
updatedAt: 2023-04-06T06:18:48Z
lastEditedAt: 2023-04-06T06:18:48Z
closedAt:
---

# github: contribution: calculate contrib to define better day for commits

Take the expected contribution of a single commit into account when choosing a day. In the original scenario the `.github` repository has an origin and four mirrors, so the author expects a published commit to raise the calendar by five contributions at once.

```text
origin # git@github.com:kamilsk/.github.git
mirror-octolab
mirror-octomation
mirror-octopot
mirror-tact
```

The intended meaning, illustrated: with a count of 8 and a target of 10, a step of 5 lands on 13, so such a day may be less suitable than one with enough headroom. Suggest must be able to explain the expected increase and its choice.

**At present** the model accounts for the current count and the target, but not for the size of the planned contribution. The number of remotes cannot be treated as a guaranteed GitHub increase: unique destinations, forks, mirrors and the publications that actually count all have to be distinguished. How the step is supplied or detected needs a separate decision.

The criterion of success: a selection that respects a confirmed or configured step, comprehensible behaviour when no suitable day exists, and compatibility with the presets of [#128](issue-128.md).
