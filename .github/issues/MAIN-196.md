---
code: MAIN-196
id: I_kwDOE2M9Zc8AAAABi0qMzw
databaseId: 6631886031
number: 568
url: https://github.com/octomation/maintainer/issues/568
title: "skills: review: ignore only the folder of private reports, not .agents/"
issueType: Bug
labels:
  - "scope: docs"
  - "scope: code"
  - "type: bug"
  - "severity: minor"
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
createdAt: 2026-11-26T17:23:37Z
updatedAt: 2026-11-26T20:01:53Z
lastEditedAt: 2026-11-26T20:01:53Z
closedAt: 2026-11-26T19:44:13Z
---

# skills: review: ignore only the folder of private reports, not .agents/

## Problem

The agents guide and the review skill tell the user to add `.agents/` to
`.gitignore` or `.git/info/exclude` before a security report. `maintainer
skill install --agent codex --project` writes the project's skills to
`.agents/skills/`, so a team that follows the advice in `.gitignore` stops
tracking its project skills, and teammates do not get them. `review.py`
itself asks only for `.agents/reviews/`.

## PoC

```bash
git init demo && cd demo
echo '.agents/' >> .gitignore           # as docs/content/agents.md advises
maintainer skill install --agent codex --project
git status --porcelain                  # only .gitignore: the skills are ignored
```

## Expected behavior

The guide and the skill ask to ignore only `.agents/reviews/`, the folder of
private reports, as `review.py` does.

## Acceptance criteria

- [x] `docs/content/agents.md` and `skills/review/SKILL.md` name
      `.agents/reviews/`, not `.agents/`, for `.gitignore` and
      `.git/info/exclude`.
- [x] The message of `review.py` and the two texts name the same folder.
- [x] `go test ./skills` and the tests of `review.py` pass.

## Out of scope

This repository's own `.gitignore`: its `.agents/` holds no project skills.

Found by the code review `.github/reports/20261126T172207Z.md` (F2).
