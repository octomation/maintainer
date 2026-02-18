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
state: CLOSED
stateReason: COMPLETED
duplicateOf: null
createdAt: 2026-11-25T18:15:34Z
updatedAt: 2026-11-26T17:26:35Z
lastEditedAt: 2026-11-26T17:25:12Z
closedAt: 2026-11-26T17:26:35Z
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

The skill:

- [x] `skills/review/` in this repository holds the skill; it is not tied to
      this repository: the checks to run, the threat model and the issue codes
      come from the reviewed project, and a project without them gets a
      stated default.
- [x] Issue codes come from the project's issue mirror with any prefix;
      without a mirror, `Tracked` takes a GitHub issue (`#N` or
      `owner/repo#N`). Plain `review.py check` stays offline. Without a
      mirror, `status` and `check --published` read the issue state with
      `gh` and fail closed: without `gh`, a login or a readable issue, the
      report stays private. The tests use a fake `gh`.
- [x] Security reports go to `.agents/reviews/` by default, and `review.py`
      refuses to write there when Git does not ignore the folder.
- [x] The description triggers on a review report or on `maintainer:review`,
      not on a bare "code review" or "security review", and says that the
      skill writes a report in a fixed form to `.github/reviews/`.
- [x] The payload holds `SKILL.md`, `references/`, `templates/` and
      `scripts/review.py`; the HTML pages and `test_review.py` live outside
      `skills/review/`, and `ci` runs the tests.

Shipping:

- [x] Every release archive and the binary carry both skills;
      `.goreleaser.yml`, `release.mjs skill` and `make release-check` check
      both.
- [x] `maintainer skill install|status|info|show|export` handle both skills;
      `--skill` selects them, both by default. A `review` folder of another
      installer is not a copy of this skill: `install` skips it with a note
      and exit `0`, and fails only for an explicit `--skill review`.
- [x] `skill status` finds the plugin by the tool name, not by the skill
      name. A plugin copy must hold exactly the binary's skills, so `--exact`
      fails without `review` in the plugin; a user or project folder is
      judged by the skills found there. Tests cover a plugin-only setup for
      Claude Code and Codex.
- [x] `ci` installs both skills from the binary and checks them with
      `skill status --exact`.
- [x] `cd` publishes with the action of octomation/skills-action#2, then
      reads the catalog at `maintainer--vX.Y.Z` and compares the exact set
      of skills and their digests with `skill info --json` of the release
      binary: a missing or extra skill fails.
- [x] The guide on agents describes the second skill.
- [x] After the release a user who updates the plugin gets
      `maintainer:review`; a review of this repository made with the
      published skill passes `review.py check`, and the local copy in
      `.agents/skills/review` is removed.

## Out of scope

- New review modes.
- Changes to the report form beyond what makes it independent of this
  repository.

<!-- 2026-11-26T17:26Z https://github.com/octomation/maintainer/issues/549#issuecomment-7199023958
Shipped in v0.6.0: the catalog plugin carries maintainer:review, skill status --exact passes with both skills from the plugin, and the first review made with the published skill (.github/reviews/20261126T172207Z.md) passes its check.
-->
