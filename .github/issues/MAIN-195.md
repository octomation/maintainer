---
code: MAIN-195
id: I_kwDOE2M9Zc8AAAABi0qHjA
databaseId: 6631884684
number: 567
url: https://github.com/octomation/maintainer/issues/567
title: "skills: install: name only the skills and agents that the run set up"
issueType: Bug
labels:
  - "scope: code"
  - "scope: test"
  - "type: bug"
  - "severity: minor"
  - "impact: low"
  - "effort: easy"
assignees:
  - kamilsk
milestone: null
parent: null
issueFields: {}
state: CLOSED
stateReason: COMPLETED
duplicateOf: null
createdAt: 2026-11-26T17:23:32Z
updatedAt: 2026-11-26T20:01:50Z
lastEditedAt: 2026-11-26T20:01:50Z
closedAt: 2026-11-26T19:44:39Z
---

# skills: install: name only the skills and agents that the run set up

## Problem

`maintainer skill install` ends with a hint to load the skills, and the hint
names every skill of the run for both agents: `/maintainer and /review in
Claude Code, $maintainer and $review in Codex`. It does so even when `install`
left a `review` folder of another tool alone, so the agent loads that tool's
skill, and when `--agent codex` set up Codex only.

## PoC

```bash
mkdir -p ~/.claude/skills/review && printf -- '---\nname: review\n---\n' > ~/.claude/skills/review/SKILL.md
maintainer skill install --agent claude-code,codex
# … left ~/.claude/skills/review alone …
# Start a new agent session to load the skills: /maintainer and /review in Claude Code, …
maintainer skill install --agent codex --project
# the same hint names Claude Code
```

## Expected behavior

The hint names, for each agent of the run, only the skills that the run
installed or found in place as this binary's copies.

## Acceptance criteria

- [x] With a foreign `review` folder for Claude Code, the hint does not name
      `/review` for Claude Code.
- [x] With `--agent codex`, the hint names Codex only; with
      `--agent claude-code`, Claude Code only.
- [x] Tests in `internal/command` cover both cases.

## Out of scope

What `install` writes and skips: it already leaves a foreign folder alone.

Found by the code review `.github/reports/20261126T172207Z.md` (F1).
