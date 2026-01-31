---
title: Use with coding agents
description: Install the maintainer skill for Claude Code or Codex, keep it in step with the binary, or get it from the OctoLab catalog.
---

# Teach your agent maintainer

maintainer ships an agent skill: instructions that teach Claude Code, Codex, and other coding agents to use it well. With the skill, an agent shows you the `fetch` plan and applies it only when you agree, reads conflicts in the plan instead of trusting exit code `0`, checks your `fetch.toml` with `fetch config validate`, uses `status` offline to find uncommitted and unpushed work, and never prints your tokens.

The skill covers the maintainer CLI and the workspace of checkouts it manages. It is written for one release of maintainer and travels with it. Every binary carries the skill for its own version.

## Install it from the binary

```sh
maintainer skill install --agent claude-code
maintainer skill install --agent codex
```

The skill goes to `~/.claude/skills/maintainer` for Claude Code (or `$CLAUDE_CONFIG_DIR/skills/maintainer`) and to `~/.agents/skills/maintainer` for Codex. Add `--project` to install it into the current directory instead, for one repository only. This works offline, needs no token and matches your binary exactly.

Start a new agent session. Ask for `/maintainer` in Claude Code or `$maintainer` in Codex, or just describe what you want from your checkouts: the agent loads the skill when it fits.

## Keep it in step

After upgrading maintainer, replace the skill with the new binary’s copy:

```sh
brew upgrade --formula octolab/tap/maintainer   # or: brew upgrade --cask octolab/tap/maintainer
maintainer skill install --agent claude-code,codex --replace
```

`--replace` only replaces a copy that maintainer installed and nobody changed since. Directories from other installers, symbolic links, and edited copies are left alone, with an explanation.

To see what your agents load, run:

```sh
maintainer skill status
```

It checks user and project copies for Claude Code and Codex, the plugins of both agents, and the legacy Codex skill directory. It reports for each copy whether it is compatible with your binary and whether its content matches the binary’s skill; for copies installed by maintainer, it also detects edits, and it warns when one agent loads more than one copy.

It exits with `1` when a copy is incompatible or was edited after maintainer installed it; `--exact` also fails on content that differs from the bundled skill. Use `maintainer skill status --agent claude-code,codex --exact` to require a matching copy for both agents: without `--agent`, finding no copies is not an error.

```text
$ maintainer skill status --agent claude-code
maintainer 0.2.1 ships the maintainer skill 0.2.1, sha256:…

Claude Code user: ~/.claude/skills/maintainer
  skill 0.2.1, compatible with this binary, same as this release, installed by maintainer, intact
```

`maintainer skill show` prints the skill, `maintainer skill info` its version and digest, and `maintainer skill export -o <dir>` writes it to any other directory.

## Or get it from the OctoLab catalog

The [octolab/skills](https://github.com/octolab/skills) catalog holds the skill of the latest stable maintainer release, next to the skills of other OctoLab tools. Use it if you prefer your agent’s own plugin manager.

| Agent | Install | Update |
| --- | --- | --- |
| Claude Code | `claude plugin marketplace add octolab/skills`, then `claude plugin install maintainer@octolab` | `claude plugin update maintainer@octolab` |
| Codex | `codex plugin marketplace add octolab/skills`, then `codex plugin add maintainer@octolab` | `codex plugin marketplace upgrade octolab` |
| Any, with the [skills CLI](https://skills.sh/) | `npx skills add octolab/skills --skill maintainer -a claude-code -a codex -g` | `npx skills update` |

The catalog follows the latest release, so an update can bring a skill newer than your binary. Each release is also tagged `maintainer--v0.2.1` in the catalog: `npx skills add 'octolab/skills#maintainer--v0.2.1' --skill maintainer` stays on that release. A `#maintainer--v0.2.1` suffix on `claude plugin marketplace add` pins the whole catalog.

## When versions differ

The skill starts by running `maintainer version`. If your binary is outside the range the skill was written for, the agent stops relying on its recipes, reads the skill your binary carries with `maintainer skill show`, and tells you which side to update. Binaries older than v0.2.1 have no `maintainer skill` command; the agent falls back to `--help`.

Keep one copy per agent. A plugin and a copy in `~/.claude/skills` both load in Claude Code, a plugin and a copy in `~/.agents/skills` both load in Codex, and the agent may follow the older one. `maintainer skill status` points out such duplicates.

## What the skill does not do

The skill never runs `fetch --apply` on its own: the agent shows you the plan and waits for your go-ahead. It does not push, merge, reset or delete checkouts, edit the state file, or remove push locks. It does not print or copy your tokens, and asks you to set `GITHUB_TOKEN` or a `token_file` yourself instead of passing a token on the command line. It runs the same commands you would, with your configuration and your limits.
