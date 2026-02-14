---
code: MAIN-188
id: I_kwDOE2M9Zc8AAAABigk3vQ
databaseId: 6610827197
number: 549
url: https://github.com/octomation/maintainer/issues/549
title: "skills: review: ship the review skill with the agent plugin"
issueType: Feature
labels:
  - "scope: docs"
  - "scope: code"
  - "scope: test"
  - "type: feature"
  - "impact: medium"
  - "effort: hard"
assignees:
  - kamilsk
milestone: null
parent: null
issueFields: {}
state: OPEN
stateReason: null
duplicateOf: null
createdAt: 2026-11-25T18:15:34Z
updatedAt: 2026-11-25T18:16:54Z
lastEditedAt: null
closedAt: null
---

# skills: review: ship the review skill with the agent plugin

## Motivation

The agent skill of maintainer drives the CLI. Reviews of this repository are
written with a second skill, `review`, which is on trial here and is not
published ([self-control](../self-control.md)). Users of the plugin get no
review skill: each maintainer writes reviews in a free form, and the result
is long, varies from one review to the next and does not show the scale of
the problems at a glance.

The skill gives one report form for three modes — code, issue and security —
with a verdict that follows from the table of findings and a checker for the
form. Published in the same plugin, it becomes `maintainer:review` next to
`maintainer:maintainer`.

## PoC

```bash
$ claude plugin update maintainer@octolab
# "code review" -> maintainer:review writes .github/reviews/<UTC>.md
$ python3 <skill-dir>/scripts/review.py check .github/reviews/20261125T152232Z.md
Code review: ACTION — Critical 0, High 0, Medium 19, Low 23.
Report: .github/reviews/20261125T152232Z.md
Coverage: full.
```

## Acceptance criteria

- [ ] `skills/review/` in this repository holds the skill; it is not tied to
      this repository: the checks to run, the threat model and the issue codes
      come from the reviewed project, and a project without them gets a
      stated default.
- [ ] Security reports stay out of the tracked tree by default.
- [ ] The release publishes it to the catalog with the main skill
      (octomation/skills-action#2); a user who updates the plugin gets
      `maintainer:review`.
- [ ] `maintainer skill install|status|info` and `go test ./skills` handle
      both skills; the guide on agents describes the second one.
- [ ] The tests of `review.py` run in `ci`.
- [ ] A review of this repository made with the published skill passes
      `review.py check`, and the local copy in `.agents/skills/review` is
      removed.

## Out of scope

- New review modes.
- Changes to the report form beyond what makes it independent of this
  repository.
