---
code: MAIN-191
id: I_kwDOE2M9Zc8AAAABivKvcA
databaseId: 6626127728
number: 564
url: https://github.com/octomation/maintainer/issues/564
title: "skills: maintainer: tell the plugin copy from the binary copy and state what skill status checks"
issueType: Bug
labels:
  - "scope: docs"
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
createdAt: 2026-11-26T11:56:34Z
updatedAt: 2026-11-26T15:14:38Z
lastEditedAt: 2026-11-26T15:14:38Z
closedAt: 2026-11-26T15:11:10Z
---

# skills: maintainer: tell the plugin copy from the binary copy and state what skill status checks

## Problem

The agent skill gives one way to fix a version mismatch: reinstall it from the
binary with `skill install --replace`. A user who has the skill from the
OctoLab catalog plugin then gets a second copy, and the agent loads both.
The recipe for `skill status` also says that it "flags … duplicate ones, and
exits `1` on a problem": a duplicate is only a warning, a copy that differs
from the binary fails only with `--exact`, and the skill does not name
`--exact`. The example pseudo-version `v0.5.2-0.20261107…` has a form that Go
does not produce after the tag `v0.5.2`: a build after that tag is
`v0.5.3-0.…`, sometimes with `+dirty`.

## PoC

```bash
claude plugin list | grep maintainer@octolab     # the skill comes from the plugin
maintainer version                                # newer than the skill's range
# the agent follows SKILL.md §1:
maintainer skill install --agent claude-code --replace
maintainer skill status --agent claude-code
# today: two copies, a warning, exit 0; the agent reads exit 0 as "fixed"
```

## Expected behavior

- §1 of `SKILL.md` asks `skill status` where the skill comes from. A plugin
  copy is updated with the agent's plugin manager
  (`claude plugin update maintainer@octolab`,
  `codex plugin add maintainer@octolab`); a copy from the binary is replaced
  with `skill install --replace`. The agent never installs a second copy.
- The recipe in `references/workflows.md` states what `skill status` reports,
  when it exits `1`, and that `--exact` also fails on content that differs from
  the binary's skill.
- The pseudo-version example has the form Go gives after a tag.

## Acceptance criteria

- [x] `SKILL.md` §1 has the plugin branch with both plugin commands, and the
      binary branch; no recipe installs a copy next to a plugin.
- [x] `references/workflows.md` describes duplicates as a warning, names
      `--exact` and the exit codes as `maintainer skill status --help` and
      `docs/content/agents.md` give them.
- [x] The pseudo-version example reads `v0.5.3-0.…` after `v0.5.2`, or
      equivalent for the current tag.
- [x] `go test ./skills` passes.

## Out of scope

Changes to `skill status` itself: they belong to [[MAIN-188]].

Related: [[MAIN-188]].
