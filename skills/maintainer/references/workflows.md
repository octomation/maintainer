# Workflows

Recipes for common requests. Placeholders such as `acme` and
`/Users/me/Development` are examples: ask the user for the real values.

## Set up a workspace

```sh
maintainer fetch config init
maintainer fetch config validate
```

`init` writes a documented `./fetch.toml`, or the file `--config` names, and
refuses to overwrite one without `--force`; don't pass `--force` over the
user's config. Ask the user to set `workspace.root`, the owners and the token
source, then validate. To keep the config for every directory, it belongs in
`~/.config/maintainer/fetch.toml`:

```sh
maintainer fetch config validate --config ~/.config/maintainer/fetch.toml
```

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
```

These read the calendar of the token's owner and need a token even for a
public profile. `suggest` prints a timestamp to stdout and exits `4` when no
moment fits up to now; it never creates a commit.

## Project helpers

```sh
maintainer go vanity --host go.example.org --file modules.yml build ./site
maintainer makefile build release.mk
```

`go vanity build` writes static `index.html` import pages for the modules in
`modules.yml` into the directory given, or the current one; it doesn't deploy
them. `makefile build` expands the `include` lines of each Makefile into
`dist/<name>/Makefile`. Both write files: tell the user where before running
them in their repository.

## Install and update this skill

The binary carries the skill for its own version:

```sh
maintainer skill info
maintainer skill install --agent claude-code,codex
maintainer skill status
```

`install` writes `~/.claude/skills/maintainer` for Claude Code and
`~/.agents/skills/maintainer` for Codex, or the same under the current
directory with `--project`, and leaves directories of other installers alone.
After upgrading maintainer, run it again with `--replace`. `status` lists every
copy the agents load, flags incompatible, changed and duplicate ones, and
exits `1` on a problem. `maintainer skill export -o skills` writes the skill
into any other directory, and `maintainer skill show --file references/cli.md`
prints one of its files.
