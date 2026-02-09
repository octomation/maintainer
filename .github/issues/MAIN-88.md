---
code: MAIN-88
id: I_kwDOE2M9Zc5ZfG7a
databaseId: 1501327066
number: 88
url: https://github.com/octomation/maintainer/issues/88
title: "git: stats: provide local-based statistics"
issueType: null
labels:
  - "scope: code"
assignees:
  - kamilsk
milestone: "[[MAIN-M1]]"
parent: null
issueFields: {}
state: OPEN
stateReason: null
duplicateOf: null
createdAt: 2022-12-17T10:58:56Z
updatedAt: 2026-11-20T15:42:46Z
lastEditedAt: 2026-11-20T15:42:46Z
closedAt: null
---

# git: stats: provide local-based statistics

## Motivation

The contribution commands read the GitHub profile calendar ([[MAIN-159]]). It
sees only what GitHub counts: it misses private work the profile hides, history
that never reached GitHub, and the share of one repository within the
aggregated calendar. The local Git history of the checkouts has all of it and
needs no network. The idea is to derive statistics directly from that history,
in the spirit of [git-stats](https://github.com/IonicaBizau/git-stats) and
[github-contributions-chart](https://github.com/sallar/github-contributions-chart).

## PoC

```bash
$ maintainer git stats --since=2022-01-01
# history:  default branches of 42 checkouts, merges excluded
# authors:  kamil@example.com, kamilsk@users.noreply.github.com
# period:   2022-01-01 .. 2026-11-20, author dates in UTC
# commits:  3120; by day: Mon 540, Tue 610, ...
```

## Acceptance criteria

- [ ] The output states which history it counted: the repositories, the refs,
  the authors and the period.
- [ ] Whether merge commits count and in which time zone a commit falls on its
  day are defined, documented and stated in the output.
- [ ] The command works offline and does not fetch; history that is not local
  is not claimed as counted.
- [ ] It works for one repository and for the checkouts of the workspace.
- [ ] A JSON mode gives the same numbers for further processing.

## Out of scope

- The status of checkouts (branches, uncommitted work, divergence): `status`.
- The GitHub calendar and its suggestions: the contribution commands.

<!-- 2026-11-20T21:00Z https://github.com/octomation/maintainer/issues/88#issuecomment-7043720925
Maturity: idea. There is no git command group (internal/command/root.go). Local Git is read already, but not as history: status inspects the working tree, ahead/behind and push locks of each checkout (internal/service/status/git.go), the contribution helper opens a repository with go-git (internal/command/github/contribution/helper.go), and the contribution commands count the GitHub calendar.
Proposal for v0.5.0: keep. Statistics from the local history of all checkouts complete the one picture of a contributor's work where GitHub sees nothing, and reuse the workspace and status code. Not a v0.5.0 deliverable: it stays in its milestone.
-->
