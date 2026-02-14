---
code: MAIN-187
id: I_kwDOE2M9Zc8AAAABigfAwQ
databaseId: 6610731201
number: 548
url: https://github.com/octomation/maintainer/issues/548
title: "fetch: config: do not read configuration from the working directory"
issueType: Bug
labels:
  - "scope: docs"
  - "scope: code"
  - "scope: test"
  - "type: bug"
  - "severity: major"
  - "impact: high"
  - "effort: medium"
assignees:
  - kamilsk
milestone: null
parent: null
issueFields: {}
state: OPEN
stateReason: null
duplicateOf: null
createdAt: 2026-11-25T18:10:01Z
updatedAt: 2026-11-25T18:12:51Z
lastEditedAt: 2026-11-25T18:10:38Z
closedAt: null
---

# fetch: config: do not read configuration from the working directory

## Problem

The content of the working directory can change what `fetch`, `status` and the
hub do: they take configuration from it before the user's own. The working
directory must not be a source of configuration.

## Expected behavior

- The configuration lookup has three steps: `--config <path>`; an existing
  file named by `$MAINTAINER_FETCH_CONFIG`;
  `$XDG_CONFIG_HOME/maintainer/fetch.{toml,yaml}` (fallback
  `~/.config/maintainer/`). `./fetch.toml` and `./fetch.yaml` are no longer
  read. Whether the folder is a Git work tree does not matter.
- No silent change of the plan: if `./fetch.toml` or `./fetch.yaml` exists and
  the run did not name it, one notice on stderr names the file and the three
  ways to use it (`--config`, the variable, a move to the user folder). If no
  other configuration is found, the command stops with exit `2` and that
  message instead of the single-run mode. A configuration named by
  `--config` or `$MAINTAINER_FETCH_CONFIG` gives no notice about a file in the
  working directory.
- `fetch config init` without `--config` writes the template to
  `$XDG_CONFIG_HOME/maintainer/fetch.toml` and prints the path; with
  `--config <path>` it writes there.
- `status` follows the same rule. The service follows it with the home folder
  as the base: a `fetch.toml` or `fetch.yaml` there that `service.toml` does
  not name gives a notice in its log, or stops the start with exit `2` when no
  other configuration is found. A `service.toml` that names the fetch
  configuration works as before.
- `--config=""` keeps its meaning: discovery off, the single-run mode, no
  notice.
- The note of v0.5.1 states the break and the three ways to move.

## Acceptance criteria

- [x] In a scratch folder without `.git`, with a `fetch.toml` and a user
      configuration in the XDG folder, `fetch` plans from the XDG file and
      prints the notice. The same in a Git checkout.
- [x] In a folder with `fetch.toml` and no other configuration, `fetch` exits
      `2` with the message and plans nothing; with `--config ./fetch.toml` it
      works as before.
- [x] `fetch config init` writes the user file, refuses to overwrite without
      `--force`, and the next `fetch` finds it.
- [x] With `--config=""` and a `fetch.toml` in the working directory,
      `fetch --owner <owner>` runs in the single-run mode without a notice.
- [x] With a `fetch.toml` in the working directory, a configuration named by
      `--config` or `$MAINTAINER_FETCH_CONFIG` is used without a notice.
- [x] With a `fetch.toml` in the home folder that `service.toml` does not
      name, `maintainer service run` logs the notice and starts on the user
      configuration; without another configuration it exits `2` and writes no
      label file.
- [x] `docs/content/fetch.md` (lookup order, onboarding, the adoption example),
      `docs/content/workspace.md`, `docs/content/service.md`,
      `docs/content/quickstart.md` and `docs/content/contributions.md` where
      they show the old flows, and the agent skill reference state the rule.

## Out of scope

- The repository helper keeps its `.env` for `gh`: [[MAIN-186]].
- The tokens of fetch profiles (`token_env`).

Related: [[MAIN-177]] holds the same rule for the token: the contribution
commands no longer read a `.env` file of the working directory.

Source: [the security review of 2026-11-25](../reviews/20261125T152238Z.md).
