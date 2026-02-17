# Workflows

Recipes for common requests. Placeholders such as `acme` and
`/Users/me/Development` are examples: ask the user for the real values.

## Set up a workspace

```sh
maintainer fetch config init
maintainer fetch config validate
```

`init` writes a documented `~/.config/maintainer/fetch.toml` (under
`$XDG_CONFIG_HOME` when it is set), or the file `--config` names, prints the
path, and refuses to overwrite one without `--force`; don't pass `--force`
over the user's config. Ask the user to set `workspace.root`, the owners and
the token source, then validate. A `fetch.toml` in the working directory is
not read; use it only with `--config` and the user's go-ahead.

## Plan, review, apply

```sh
maintainer fetch
maintainer fetch --format=json | jq '.summary'
maintainer fetch --format=json | jq '.actions[] | select(.kind == "conflict" or .kind == "orphan")'
```

Summarize the plan for the user: how many clones, moves, remote updates and
adoptions, and every conflict with its reason. After their go-ahead:

```sh
maintainer fetch --apply
```

Apply runs adoptions and relocations, then remote updates, then moves, then
clones and fetches. Every adopted, relocated, re-pointed or moved checkout is
fetched in the same run. Exit `3` means some repositories failed or stayed in
conflict; the rest is applied.

## Onboard an existing tree

For checkouts already laid out as `<root>/{public,private,internal}/<owner>/<repo>`:

```sh
maintainer fetch --config ~/.config/maintainer/fetch.toml
maintainer fetch --config ~/.config/maintainer/fetch.toml --apply
maintainer fetch --config ~/.config/maintainer/fetch.toml --apply
maintainer fetch state show | jq '.repos | length'
```

The first apply adopts existing checkouts and clones missing repositories; the
second moves checkouts that were adopted at the old path of a renamed
repository. From then on, a run against unchanged repositories only fetches.
`[filters]` such as `exclude_forks` skip new clones and adoptions only.

## Narrow a run

```sh
maintainer fetch --owner acme
maintainer fetch --profile bot --format=json
maintainer fetch --owner acme --timeout 5m --apply
```

A repository outside the selection gets no clone, fetch, move or remote update;
JSON lists a tracked one as a `noop` with the reason `outside the selection`.
An id-only `[[repos]]` pin is still confirmed once by id, and managed only when
its owner is selected.

## One run without a config

```sh
maintainer fetch --config="" --owner acme
```

This plans under the current directory with the default layout and the
default state file; the token comes from `GITHUB_TOKEN`.

## Find local work

```sh
maintainer status --format=plain
maintainer status --owner acme --format=json | jq '.[] | select(.changed_files > 0 or .untracked > 0) | {repository, path}'
maintainer status --format=json | jq -c '.[] | select(.ahead > 0) | {repository, branch, ahead}'
maintainer status --format=json | jq -c '.[] | select(.push_locked) | {repository, path}'
maintainer status --root ~/Development --format=plain
```

status reads no network: ahead and behind come from the last fetch. A row
with `error` failed to inspect; the others are complete. `--root` overrides
`workspace.root` for one run.

## Keep a checkout outside the tree

To keep a checkout where it is, such as `~/.dotfiles`, the user adds a
per-repository rule by id:

```toml
[[repos]]
match = { id = 154873464 }
path = "~/.dotfiles"
```

fetch verifies the identity and adopts the path on the next apply; a copy at
the layout path becomes an `orphan` and stays untouched. A tree of checkouts
that should never move goes to `workspace.pins` instead.

## Tidy the state

```sh
maintainer fetch state show | jq '.repos'
maintainer fetch state prune
```

`prune` forgets records whose directory no longer exists and saves the state
at once, without `--apply`; checkouts themselves are never touched. Run it
only when the user asks.

## Plan a contribution

```sh
maintainer github contribution lookup now/-3
maintainer github contribution suggest --short
maintainer github contribution histogram 2026-10
maintainer github contribution snapshot 2026 > snap.json
maintainer github contribution diff snap.json 2026
git commit --date="$(maintainer github contribution suggest --short)"
```

These read the calendar of the token's owner and need a token even for a
public profile. They take it only from `GITHUB_TOKEN` in the environment,
never from a `fetch.toml` profile. The days match the profile's calendar of
the year in any time zone, and a repeated run reads only the months that
changed, from the cache in `$XDG_CACHE_HOME/maintainer/contributions` (or
`~/.cache`). Without a date, `suggest` anchors at the author date of `HEAD` in
the current repository (now outside one or before its first commit), prints a
UTC timestamp to stdout for `git commit --date`, and never creates a commit
itself; run that commit only with the user's go-ahead. It exits `2` before
any request without a token or for an anchor in the future, and `4` when no
moment fits up to now. A `warn:` line names the months whose reviews may be a
day off because GitHub gives no date for some of their contributions.

## Install and update this skill

The binary carries this skill and the review skill for its own version:

```sh
maintainer skill info
maintainer skill status
maintainer skill install --agent claude-code,codex
```

`status` lists every copy of both skills that the agents load: the user's
and the project's skills of Claude Code and Codex, their `maintainer`
plugins and the legacy Codex directory. For each copy it reports whether it
is compatible with the binary, whether it is the same as this release, and
who installed it; a plugin copy says `plugin: maintainer@<catalog>`. It exits
`1` when a copy is incompatible, was changed after maintainer installed it,
or, for a requested `--agent`, is missing; with `--exact`, also when a copy
differs from this release or a plugin does not hold both skills. Two copies
of one skill for one agent are only a warning, and so is a missing review
skill outside a plugin. Usage errors exit `2`.

Update a copy where it came from. A plugin copy is updated by the agent's
plugin manager: `claude plugin update maintainer@octolab`, or
`codex plugin marketplace upgrade octolab` and then
`codex plugin add maintainer@octolab`. A copy installed by maintainer is
replaced after upgrading the binary with `install --replace`, which writes
`~/.claude/skills/<skill>` for Claude Code and `~/.agents/skills/<skill>` for
Codex, or the same under the current directory with `--project`. Don't
install from the binary next to a plugin: the agent loads both. `install`
leaves directories of other installers alone and skips a review skill of
another tool; `--skill` picks one skill. `maintainer skill export -o skills`
writes `skills/maintainer` and `skills/review` into any other directory, and
`maintainer skill show --file references/cli.md` prints one of this skill's
files.
