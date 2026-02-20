---
code: MAIN-177
id: I_kwDOE2M9Zc8AAAABigenTA
databaseId: 6610724684
number: 543
url: https://github.com/octomation/maintainer/issues/543
title: "github: contribution: take the token only from GITHUB_TOKEN in the environment"
issueType: Bug
labels:
  - "scope: code"
  - "scope: test"
  - "type: bug"
  - "severity: major"
  - "impact: medium"
  - "effort: easy"
assignees:
  - kamilsk
milestone: null
parent: null
issueFields: {}
state: CLOSED
stateReason: COMPLETED
duplicateOf: null
createdAt: 2026-11-25T18:09:38Z
updatedAt: 2026-11-25T20:23:58Z
lastEditedAt: 2026-11-25T20:23:58Z
closedAt: 2026-11-25T18:42:42Z
---

# github: contribution: take the token only from GITHUB_TOKEN in the environment

## Problem

The contribution commands can take a token from the content of the working
directory, so a folder decides whose calendar they show. The token must come
only from the environment of the user who runs the command.

## Expected behavior

The contribution commands take the token only from `GITHUB_TOKEN` in the
environment. A `.env` file in the working directory is not read. `--token` is
not a way to pass the token: it is retired by [[MAIN-183]]. Without
`GITHUB_TOKEN`, the commands stop with the error of a missing token, exit `2`,
and send no request.

## Acceptance criteria

- [x] In a folder with a `.env` that holds `GITHUB_TOKEN` and with no token in
      the environment, `maintainer github contribution lookup` exits `2` with
      the error of a missing token and sends no request.
- [x] A test runs a contribution command in a folder with a `.env` and without
      `GITHUB_TOKEN`, and expects that error.
- [x] With `GITHUB_TOKEN` in the environment, the commands work as the guides
      describe, and the guides name no other source of the token.

## Out of scope

- The tokens of fetch profiles (`token_env`).
- The repository helper keeps its `.env` for `gh`: [[MAIN-186]].

Related: [[MAIN-187]], the same rule for the fetch configuration.

Source: [20261125T152232Z](../reports/20261125T152232Z.md), "Contribution commands take a token from a `.env` file in the working directory".
