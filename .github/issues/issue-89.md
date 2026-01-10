---
code:
id: I_kwDOE2M9Zc5ZfOVU
databaseId: 1501357396
number: 89
url: https://github.com/octomation/maintainer/issues/89
title: "git: config: contribution since"
labels:
  - "scope: code"
milestone: "[[milestone-1]]"
state: CLOSED
stateReason: NOT_PLANNED
createdAt: 2022-12-17T12:04:38Z
updatedAt: 2026-09-25T05:14:16Z
lastEditedAt: 2026-09-25T05:14:16Z
closedAt: 2023-03-25T20:55:08Z
---

# git: config: contribution since

Let the lower contribution date be set per project, so it does not have to be supplied on every call of a Git wrapper — avoiding that manual job. Projects start at different moments: Tact, for example, from 1 May 2022 and the others from 2021.

The historical prototype:

```bash
git config contribution.since 2022-05-01
# A later wrapper call uses this bound.
```

The precedence of the project setting, an explicitly passed date, and the date of the latest commit has to be defined. Choosing an earlier start must not silently cancel the chronology requirement for new commits.

Related: the automatic Git anchor [[issue-58]], the ordering of dates [[issue-84]].
