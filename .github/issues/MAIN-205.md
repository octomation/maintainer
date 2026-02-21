---
code: MAIN-205
id: I_kwDOE2M9Zc8AAAABjEhrxg
databaseId: 6648523718
number: 583
url: https://github.com/octomation/maintainer/issues/583
title: "status: bound git status of a slow checkout and show its counts as unknown"
issueType: Bug
labels:
  - "effort: medium"
  - "impact: medium"
  - "scope: code"
assignees:
  - kamilsk
milestone: null
parent: null
issueFields: {}
state: CLOSED
stateReason: COMPLETED
duplicateOf: null
createdAt: 2026-11-27T14:36:40Z
updatedAt: 2026-11-27T18:39:20Z
lastEditedAt: 2026-11-27T18:39:20Z
closedAt: 2026-11-27T18:14:35Z
---

# status: bound git status of a slow checkout and show its counts as unknown

## Problem

v0.6.1 bounds rename detection in one checkout, but past the bound `status`
runs `git status --no-renames` again without any bound. In a checkout with a
very large tree, `git status` itself takes minutes: on the owner's workspace,
one checkout of 23.5 GB took 87–90 s of a 98.7 s run over 168 checkouts, in
every variant of `git status`. The run and the hub wait for it, which the
v0.6.1 note promised would not happen.

## Expected behavior

`git status` of one checkout has a bound. Past it, the row keeps what cheaper
commands give, the counts that need `git status` are unknown, and the run goes
on.

## Acceptance criteria

- [x] A test with a `git status` that never returns finishes in about twice
      the bound.
- [x] Past the bound, the row keeps repository, default branch, branch,
      divergence and push lock, read by commands that do not walk the tree.
- [x] Past the bound, the file counts and the line counts of the row show `?`
      in the plain output; the JSON gives `null` for the changed, untracked
      and conflicted counts and for the line counts; the row counts as
      uncommitted work and does not make the run exit 1.
- [x] Plain output, JSON and the hub `/status` show the same row.
- [x] `docs/content/status.md` and `skills/maintainer/references/output.md`
      state the bound and how the row shows it.

## Out of scope

Making `git status` faster. A configurable bound.
