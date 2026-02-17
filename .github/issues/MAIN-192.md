---
code: MAIN-192
id: I_kwDOE2M9Zc8AAAABivK0FA
databaseId: 6626128916
number: 565
url: https://github.com/octomation/maintainer/issues/565
title: "skills: maintainer: give the token, anchor and exits of the contribution commands"
issueType: Bug
labels:
  - "scope: docs"
  - "scope: test"
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
createdAt: 2026-11-26T11:56:39Z
updatedAt: 2026-11-26T15:14:41Z
lastEditedAt: 2026-11-26T15:14:41Z
closedAt: 2026-11-26T14:39:15Z
---

# skills: maintainer: give the token, anchor and exits of the contribution commands

## Problem

The agent skill has one rule for tokens: "ask the user to export
`GITHUB_TOKEN` or set `token_file`". `token_file` is a key of a `fetch.toml`
profile; the `github contribution` commands read the token only from
`GITHUB_TOKEN` in the environment. An agent that follows the rule sets a
`token_file` and the command still fails. The recipe also omits the default
anchor of `suggest` (the date of `HEAD` of the current repository), the main
use (`git commit --date`), and exit `2` for a future anchor or a missing token.

A second statement is wrong: "`fetch` flags such as `--config`, `--owner`,
`--format` and `--timeout` work for its subcommands too".
`fetch config validate --format=json` prints text, and `fetch state show`
prints JSON whatever `--format` says.

## PoC

```bash
unset GITHUB_TOKEN
maintainer github contribution suggest --short
# exit 2; the skill sends the agent to token_file, which this command ignores
maintainer fetch config validate --format=json
# ok: … — not JSON, as the skill implies
```

## Expected behavior

- `SKILL.md` §7 and the contribution recipe say that these commands take the
  token from `GITHUB_TOKEN` only, and `fetch` from its profile
  (`token_file`, `token_env`, then `GITHUB_TOKEN`).
- The contribution recipe names the default anchor, the `git commit --date`
  use, and the exits `2` and `4`.
- The statement on subcommand flags names only the flags that the
  subcommands take, as their `--help` lists them.

## Acceptance criteria

- [x] The skill names `GITHUB_TOKEN` as the only token source of
      `github contribution`, and `token_file` only for `fetch`.
- [x] The recipe has an example
      `git commit --date="$(maintainer github contribution suggest --short)"`
      and states exits `2` and `4`, as `--help` and
      `docs/content/contributions.md` give them.
- [x] The subcommand statement matches `--help` of `fetch config validate`,
      `fetch state show` and `fetch state prune`.
- [x] `go test ./skills` passes.

## Out of scope

Changes to how the commands find the token.

Related: [[MAIN-188]], [[MAIN-191]].
