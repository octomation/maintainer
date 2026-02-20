---
code: MAIN-202
id: I_kwDOE2M9Zc8AAAABjEVsNA
databaseId: 6648327220
number: 580
url: https://github.com/octomation/maintainer/issues/580
title: "skills: review: write reports to .github/reports/ and .agents/reports/"
issueType: Task
labels:
  - "effort: medium"
  - "impact: medium"
  - "scope: docs"
  - "scope: test"
assignees:
  - kamilsk
milestone: null
parent: null
issueFields: {}
state: OPEN
stateReason: null
duplicateOf: null
createdAt: 2026-11-27T14:24:44Z
updatedAt: 2026-11-27T18:03:55Z
lastEditedAt: 2026-11-27T18:03:54Z
closedAt: null
---

# skills: review: write reports to .github/reports/ and .agents/reports/

## Motivation

The review skill writes its reports to `.github/reviews/`, while documents of
the same kind live next to it in `.github/reports/`. A repository ends up with
two folders for one genre, and the private twin `.agents/reviews/` clashes
with other tools that keep their own review scratch files there. The skill
calls its output a report: the folders should say so too.

## Scope

- `review.py new` writes code and issue reports to `.github/reports/` and
  security reports to `.agents/reports/`; the `Publish:` line moves a report
  from `.agents/reports/` to `.github/reports/` under the same name.
- `status`, `check` and `check --published` read only these folders, and a
  `Source:` line links `../reports/` or a URL under `.github/reports/`. The
  folders of v0.6.x are not read: a repository moves its reports once.
- `new security` refuses to write while Git does not ignore
  `.agents/reports/`, and says which line to add.
- SKILL.md, its references and the agents guide name the new folders.
- This repository moves its four reports to `.github/reports/` and its
  `Source:` links to `../reports/`.

## Acceptance criteria

- [x] `review.py new code|issue` prints a path in `.github/reports/`, `new
      security` one in `.agents/reports/`.
- [x] `status` and `check --published` read `.github/reports/` and
      `.agents/reports/`, and a `Source:` link to `../reports/`.
- [x] `check` refuses a security report in `.github/reports/` before
      publication, and a report outside the two folders.
- [x] The `Publish:` line moves a private report from `.agents/reports/` to
      `.github/reports/`.
- [x] The tests of `review.py` cover the new folders; `go test ./skills`
      passes.
- [x] `docs/content/agents.md` asks to ignore `.agents/reports/`.
- [x] `.github/reviews/` of this repository is gone: its reports are in
      `.github/reports/`, the mirror links `../reports/`, and `review.py
      status` reports no error.

## Out of scope

Moving the `.github/reviews/` folders of other repositories: their content is
not in the form of the skill. The form of a report.
