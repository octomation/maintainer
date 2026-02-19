---
code: MAIN-201
id: I_kwDOE2M9Zc8AAAABi08NRA
databaseId: 6632181060
number: 579
url: https://github.com/octomation/maintainer/issues/579
title: "skills: review: say how to treat a large target, test fixtures and Excluded"
issueType: Task
labels:
  - "effort: easy"
  - "impact: low"
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
createdAt: 2026-11-26T17:42:35Z
updatedAt: 2026-11-26T17:42:35Z
lastEditedAt: null
closedAt: null
---

# skills: review: say how to treat a large target, test fixtures and Excluded

## Problem

The first review made with the published `maintainer:review`, the code review
of v0.6.0, left three places where the skill gives no rule, and the reviewer
had to choose:

- The target that the user names explicitly is larger than one pass. The skill
  marks a lens `partial`, but does not say whether to narrow the target, split
  it into passes or stop.
- A test file under `testdata/`, such as a test of a script, is a fixture or a
  test. The Tests lens counts it either way.
- `Excluded` lists folders that the skill itself leaves out, next to folders
  that the request leaves out, in the same words.

## Expected behavior

`skills/review/SKILL.md` or its references state one rule for each case, and a
report made by the skill follows it.

## Acceptance criteria

- [x] The skill states what to do when the named target is larger than one
      pass.
- [x] The skill states how the Tests lens counts test files under `testdata/`.
- [x] The skill states how `Excluded` tells the skill's own exclusions from
      those of the request, and `review.py check` accepts the form.
- [x] `go test ./skills` and the tests of `review.py` pass.

## Out of scope

The lenses, severities and verdicts of the skill.

Source: [the code review of v0.6.0](../reviews/20261126T172207Z.md).
