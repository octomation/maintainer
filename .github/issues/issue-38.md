---
code:
id: I_kwDOE2M9Zc5JpTBA
databaseId: 1235562560
number: 38
url: https://github.com/octomation/maintainer/issues/38
title: "github: contribution: lookup shows incorrect scope for 1 week with now ts"
labels:
  - "scope: code"
milestone: "[[milestone-1]]"
state: CLOSED
stateReason: COMPLETED
createdAt: 2022-05-13T18:34:22Z
updatedAt: 2022-06-15T10:15:49Z
lastEditedAt:
closedAt: 2022-05-21T10:52:44Z
---

# github: contribution: lookup shows incorrect scope for 1 week with now ts

Fix which weeks are selected when the calendar is viewed relative to the current date. In the original case `/1` showed the previous week instead of the current one, and `/2` added a superfluous older week.

The observation from May 2022:

```text
lookup /1: showed #18 (1–7 May); #19 was wanted.
lookup /2: showed #17, #18, #19; #17 is superfluous.
```

The user must see exactly the window around the chosen date, with no unexpected shift into the past. A test has to pin the run date and cover a single week, several weeks, and Sunday separately.

**Current state:** this historical issue is closed, but the modern syntax distinguishes a centred window `/N` from a directed offset `/+N` or `/-N`; a negative value includes the reference week and the ones before it. For the present moment `now` has to be spelled out, since an empty date may mean HEAD. The direction contract is covered by [#41](issue-41.md); the crash on an argument without a suffix is [#155](issue-155.md).

<!-- 2022-05-21T10:52Z https://github.com/octomation/maintainer/issues/38#issuecomment-1133599244
https://github.com/octomation/maintainer/commit/7b3c4af8d384810c8457bd168fe5232dec9dc96e
-->
