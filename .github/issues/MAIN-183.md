---
code: MAIN-183
id: I_kwDOE2M9Zc8AAAABigesLA
databaseId: 6610725932
number: 544
url: https://github.com/octomation/maintainer/issues/544
title: "github: token: take the token only from the environment"
issueType: Bug
labels:
  - "scope: docs"
  - "scope: code"
  - "scope: test"
  - "type: bug"
  - "severity: major"
  - "impact: high"
  - "effort: easy"
assignees:
  - kamilsk
milestone: null
parent: null
issueFields: {}
state: CLOSED
stateReason: COMPLETED
duplicateOf: null
createdAt: 2026-11-25T18:09:42Z
updatedAt: 2026-11-25T19:39:34Z
lastEditedAt: 2026-11-25T19:39:34Z
closedAt: 2026-11-25T18:40:33Z
---

# github: token: take the token only from the environment

## Problem

A GitHub token given on the command line is visible to other accounts of the
same machine. The token must reach the tool only through the environment.

## Expected behavior

`--token` is retired in this release, not deprecated. The parser still knows the
flag but `--help` does not list it. Any argument that is `--token` or starts
with `--token=`, with any value or an empty one, anywhere on the command line,
also next to `--help` or `-h` and after `--`, stops the run before any request
with a usage error (exit `2`) that says why the flag is gone and to set
`GITHUB_TOKEN`. The value is neither used nor printed.

Every guide, the README, the agent skill and the error texts name only
`GITHUB_TOKEN`, and `token_env` of a profile for `fetch`. The note of v0.5.1
states the break and the replacement.

## Acceptance criteria

- [x] `maintainer fetch --help` and `maintainer github --help` do not list
      `--token`.
- [x] `maintainer fetch --token=x` and
      `maintainer github --token=x contribution lookup` exit `2`, name
      `GITHUB_TOKEN`, and send no request.
- [x] Next to `--help` or `-h`, as in `maintainer fetch --token=x --help` or
      `maintainer -h fetch --token x`, the run exits `2` with the same error;
      so does `--token=x` after `--`.
- [x] `grep -rn -- '--token' README.md docs/content skills/maintainer` finds
      only published notes of earlier releases and the text of the removal.
- [x] The message for a missing token names `GITHUB_TOKEN` only.
- [x] The note of v0.5.1 states the break and the replacement.

## Out of scope

- Published notes of earlier releases stay as they are.
- Reading a `.env` file of the working directory: [[MAIN-177]].

Source: [the security review of 2026-11-25](../reviews/20261125T152238Z.md), F3.
