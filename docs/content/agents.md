---
title: Use with coding agents
description: Install the maintainer and review skills for Claude Code or Codex, keep them in step with the binary, or get them from the OctoLab catalog.
---

# Teach your agent maintainer

maintainer ships an agent skill: instructions that teach Claude Code, Codex, and other coding agents to use it well. With the skill, an agent shows you the `fetch` plan and applies it only when you agree, reads conflicts in the plan instead of trusting exit code `0`, checks your `fetch.toml` with `fetch config validate`, uses `status` offline to find uncommitted and unpushed work, and never prints your tokens.

The skill covers the maintainer CLI and the workspace of checkouts it manages. It is written for one release of maintainer and travels with it. Every binary carries the skill for its own version, together with a second skill, [review](#write-review-reports), which writes review reports of a repository.

## Install it from the binary

```sh
maintainer skill install --agent claude-code
maintainer skill install --agent codex
```

Both skills go to `~/.claude/skills/maintainer` and `~/.claude/skills/review` for Claude Code (or under `$CLAUDE_CONFIG_DIR/skills`) and to `~/.agents/skills/maintainer` and `~/.agents/skills/review` for Codex. Add `--project` to install them into the current directory instead, for one repository only, and `--skill maintainer` or `--skill review` to install one of them. This works offline, needs no token and matches your binary exactly.

A `review` folder that another tool installed stays as it is: `install` skips it with a note and installs the maintainer skill. Only `--skill review` fails on it, as on any folder of another installer.

If you use the [catalog plugin](#or-get-it-from-the-octolab-catalog), don't install from the binary as well: the agent would load both copies.

Start a new agent session. Ask for `/maintainer` in Claude Code or `$maintainer` in Codex, or just describe what you want from your checkouts: the agent loads the skill when it fits.

## Keep it in step

After upgrading maintainer, replace the skills with the new binary’s copies:

```sh
brew upgrade --formula octolab/tap/maintainer
maintainer skill install --agent claude-code,codex --replace
```

`--replace` only replaces a copy that maintainer installed and nobody changed since. Directories from other installers, symbolic links, and edited copies are left alone, with an explanation.

To see what your agents load, run:

```sh
maintainer skill status
```

It checks user and project copies of both skills for Claude Code and Codex, the `maintainer` plugins of both agents, and the legacy Codex skill directory. It reports for each copy whether it is compatible with your binary and whether its content matches the binary’s skill; for copies installed by maintainer, it also detects edits. It warns when one agent loads more than one copy of a skill, notes a plugin that lacks the review skill, and leaves a `review` folder of another tool unchecked.

A binary built from an untagged commit, whose version is `dev` or a pseudo-version such as `v0.6.2-0.20261126121749-fbe59efd127f`, sometimes with `+dirty` at the end, names no release: `skill status` reports `compatibility unknown` for each copy and does not fail on it.

It exits with `1` when a copy is incompatible or was edited after maintainer installed it. `--exact` also fails on content that differs from the bundled skills and on a plugin that does not hold exactly the binary’s skills. Use `maintainer skill status --agent claude-code,codex --exact` to require a matching maintainer skill for both agents: without `--agent`, finding no copies is not an error. A missing review skill outside a plugin is a note; `--skill review` requires it too.

```text
$ maintainer skill status --agent claude-code
maintainer 0.6.1 ships the maintainer skill 0.6.1, sha256:…
maintainer 0.6.1 ships the review skill 0.6.1, sha256:…

Claude Code plugin: maintainer@octolab ~/.claude/plugins/cache/octolab/maintainer/0.6.1/skills/maintainer
  skill 0.6.1, compatible with this binary, same as this release, installed by another tool

Claude Code plugin: maintainer@octolab ~/.claude/plugins/cache/octolab/maintainer/0.6.1/skills/review
  skill 0.6.1, compatible with this binary, same as this release, installed by another tool
```

`maintainer skill show` prints the maintainer skill and `--skill review` the other one, `maintainer skill info` gives the version and digest of each, and `maintainer skill export -o <dir>` writes both to any other directory, as `<dir>/maintainer` and `<dir>/review`.

## Write review reports

The review skill writes one review report of a repository in a fixed form, in one of three modes: a code review asks whether you can release the code, an issue review whether the backlog states the truth, and a security review who can harm your users or your release channel. Each report has a table of findings with a severity and a verdict that follows from it, and the skill checks the form with its own script before it hands the report over. It takes the checks to run, the threat model and the issue codes from your project, and states the default it used where the project names none.

Ask for a review report, such as “write a code review report”, or for `maintainer:review` with the catalog plugin. With the binary’s copy, ask for `/review` in Claude Code or `$review` in Codex. The skill needs `python3` and Git; without an issue mirror in `.github/issues/`, it reads the state of your GitHub issues with a signed-in `gh`.

Code and issue reports go to `.github/reports/`. A security report is private: it goes to `.agents/reports/`, and the skill refuses to write it there unless Git ignores the folder, so add `.agents/reports/` to `.gitignore` or `.git/info/exclude` first. You publish a private report yourself once each of its findings is closed: the skill prints the command that moves it to `.github/reports/`.

The skill reads only these two folders.

## Or get it from the OctoLab catalog

The [octolab/skills](https://github.com/octolab/skills) catalog holds the skills of the latest stable maintainer release in one plugin, `maintainer`, next to the skills of other OctoLab tools: the agents load them as `maintainer:maintainer` and `maintainer:review`. Use it if you prefer your agent’s own plugin manager.

| Agent | Install | Update |
| --- | --- | --- |
| Claude Code | `claude plugin marketplace add octolab/skills`, then `claude plugin install maintainer@octolab` | `claude plugin update maintainer@octolab` |
| Codex | `codex plugin marketplace add octolab/skills`, then `codex plugin add maintainer@octolab` | `codex plugin marketplace upgrade octolab`, then `codex plugin add maintainer@octolab` |
| Any, with the [skills CLI](https://skills.sh/) | `npx skills add octolab/skills --skill maintainer -a claude-code -a codex -g` | `npx skills update` |

The catalog follows the latest release, so an update can bring a skill newer than your binary. Each release is also tagged `maintainer--v0.6.1` in the catalog: `npx skills add 'octolab/skills#maintainer--v0.6.1' --skill maintainer` stays on that release. A `#maintainer--v0.6.1` suffix on `claude plugin marketplace add` pins the whole catalog.

## When versions differ

The skill starts by running `maintainer version`. If your binary is outside the range the skill was written for, the agent stops relying on its recipes, reads the skill your binary carries with `maintainer skill show`, and tells you which side to update. Binaries older than v0.2.1 have no `maintainer skill` command; the agent falls back to `--help`.

The agent asks `maintainer skill status` where the skill comes from and updates it there: a plugin copy with the agent’s plugin manager, a copy from the binary with `maintainer skill install --replace`.

Keep one copy per agent. A plugin and a copy in `~/.claude/skills` both load in Claude Code, a plugin and a copy in `~/.agents/skills` both load in Codex, and the agent may follow the older one. `maintainer skill status` points out such duplicates.

## What the skill does not do

The skill never runs `fetch --apply` on its own: the agent shows you the plan and waits for your go-ahead. It does not push, merge, reset or delete checkouts, edit the state file, or remove push locks. It does not print or copy your tokens, and asks you to set `GITHUB_TOKEN` or a `token_file` yourself instead of passing a token on the command line. It runs the same commands you would, with your configuration and your limits.
