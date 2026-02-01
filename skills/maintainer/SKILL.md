---
name: maintainer
description: >-
  Manage a workspace of local GitHub checkouts with the maintainer CLI: plan
  and apply fetch (clone, refresh, move, adopt), check fetch.toml against its
  schema, read the state and pins, and see branches, uncommitted work,
  divergence and push locks across checkouts with status. Also runs the local
  hub, maintainer service, that shows the checkouts and the issues they mirror
  in one place in the browser, and plans contributions with the GitHub
  contribution calendar. Use when the user wants to clone or refresh their
  repositories, find local work across checkouts, run or configure the local
  hub, or troubleshoot a maintainer command. Not for mirroring GitHub issues into
  .github/.
license: MIT
compatibility: >-
  Requires the maintainer binary >=0.3.0 <0.4.0 on macOS or Linux and Git on
  PATH. fetch and the contribution commands need network access and a GitHub
  token; status works offline.
metadata:
  author: octomation
  version: "0.3.0"
  tool: maintainer
  tool-version-range: ">=0.3.0 <0.4.0"
---

# maintainer

maintainer is a macOS/Linux CLI that keeps a tree of local GitHub checkouts in
step with GitHub. `maintainer fetch` discovers the user's repositories and
plans what to clone, refresh, move or adopt, like `terraform plan`; `--apply`
carries the plan out without touching branches or working files.
`maintainer status` shows, offline, where local work needs attention. Both read
one config, `fetch.toml`, and one state file. `maintainer service run` brings
the checkouts and the issues they mirror from GitHub into one local hub in the
browser.

## 1. Check the binary first

This skill describes maintainer `>=0.3.0 <0.4.0`. Before anything else:

1. Run `maintainer version` and read the `version` line.
2. Compare it with that range.
   - In range: continue, and read `maintainer <command> --help` before using a
     command's flags.
   - Older or newer: don't rely on the recipes here. Read the skill that ships
     with the binary, `maintainer skill show`, and tell the user the versions
     differ; suggest upgrading maintainer
     (`brew upgrade --formula octolab/tap/maintainer`) or reinstalling the
     skill from the binary
     (`maintainer skill install --agent claude-code --replace`, or `codex`).
   - `dev`, a pseudo-version such as `v0.3.0-0.20261021…`, or unknown: treat
     compatibility as unknown and rely on `--help`.
   - `maintainer skill` missing: the binary predates it; rely on `--help` only.
   - `maintainer` not found: explain how to install it
     (https://maintainer.octolab.org/quickstart/); don't install it without
     the user's go-ahead.
3. Never guess a flag or a config key that `--help` or the schema doesn't show.

## 2. Ground rules

- **Plan first; apply only with the user's go-ahead.** `maintainer fetch`
  without `--apply` changes no checkout and saves no repository state. Show
  the user the plan, in particular `conflict`, `move`, `relocate`,
  `update_remote` and `orphan` rows, and run `maintainer fetch --apply` only
  when they agree. Each `--apply` builds a fresh plan; it never replays a
  plan printed earlier.
- **Tokens stay secret.** Never print `GITHUB_TOKEN`, a `token_file` or an
  inline `token`, and don't put a token on the command line with `--token`:
  it stays in shell history. Ask the user to export the variable or set
  `token_file` themselves.
- **Leave the checkouts to Git and the user.** maintainer never pushes,
  merges, resets or deletes, and neither should you on its behalf: don't
  delete an `orphan`, don't remove a push lock (`no_push`), don't edit the
  state file by hand. Removing a directory, such as an incomplete clone, is
  the user's call.
- **status is offline and read-only.** Its ahead and behind counts come from
  local remote-tracking refs; refresh them with `maintainer fetch --apply`,
  with the go-ahead above.
- **Don't open the interactive table.** In a terminal `maintainer status`
  starts a TUI; use `--format=plain` or `--format=json`.
- **Start narrow.** `--owner` and `--profile` limit a fetch run, and
  `--timeout 5m` bounds it; widen only when the user asks.
- **Exit code 0 doesn't mean nothing to do.** A plan with conflicts exits `0`:
  check `summary.conflict` (see [references/output.md](references/output.md)).

## 3. Choose the command

| The user wants | Command |
| --- | --- |
| A first config | `fetch config init` (writes `./fetch.toml`; `--force` overwrites) |
| To check a config | `fetch config validate`, or `cue vet -c fetch.cue fetch.toml` against the schema |
| To see what fetch would do | `fetch`, or `fetch --format=json` |
| To clone missing and refresh existing checkouts | `fetch --apply`, after reviewing the plan |
| One owner or one account only | `fetch --owner acme` or `fetch --profile bot` |
| Work in progress across checkouts | `status --format=plain` or `status --format=json` |
| Checkouts of some owners | `status --owner acme --format=json` |
| What fetch remembers | `fetch state show` |
| To forget checkouts whose directory is gone | `fetch state prune` (writes the state at once) |
| A time to contribute | `github contribution suggest`; also `contribution lookup`, `contribution histogram`, `contribution snapshot`, `contribution diff` |
| Go vanity import pages | `go vanity build` |
| Combined Makefiles | `makefile build` |
| maintainer in a browser on this computer | `service run`, or `brew services start octolab/tap/maintainer` |

`fetch` flags such as `--config`, `--owner`, `--format` and `--timeout` work
for its subcommands too. Recipes:
[references/workflows.md](references/workflows.md). Every command and flag of
this release: [references/cli.md](references/cli.md).

## 4. Configuration, state and tokens

- **Lookup**, first match wins: `--config <path>` (`--config=""` disables
  discovery), the file named by `$MAINTAINER_FETCH_CONFIG`, `./fetch.toml`,
  `./fetch.yaml`, then `$XDG_CONFIG_HOME/maintainer/fetch.toml` or `.yaml`
  (usually `~/.config/maintainer/`). Without a config, `fetch` needs an
  `--owner` and works under the current directory.
- **The schema is the contract.** The keys are `[workspace]` (`root`, `path`,
  `pins`), `[defaults]` (`clone_url`, `concurrency`, `state_file`),
  `[filters]`, `[profiles.<name>]`, `[[owners]]` and `[[repos]]`; the schema
  `fetch.cue` lives in the maintainer repository under `internal/config/`. An
  unknown key or table is bad config: fetch, its subcommands and status exit
  `2` and suggest the key it was probably meant to be. Fix the key; don't
  remove what you don't understand.
- **Layout is scope.** `workspace.path`, by default
  `{{.Visibility}}/{{.Owner}}/{{.Repo}}`, is both where new clones go and
  the only place both commands look. Checkouts elsewhere are managed only
  through `workspace.pins` or a per-repository `path` in `[[repos]]`; match
  `[[repos]]` by `id` so a rule survives a rename.
- **Profiles and tokens.** A profile is a token with the owners it discovers
  (`include_owners`, `["*"]` for the user and every organization they belong
  to). Its token comes from `token_file` (mode `0600` at most), then
  `token_env` (`GITHUB_TOKEN` for the first profile in alphabetical order),
  then an inline `token`. fetch needs a token even for public repositories;
  status needs none.
- **State.** `$XDG_STATE_HOME/maintainer/fetch/state.json` (usually
  `~/.local/state/maintainer/fetch/`), or `defaults.state_file`. It is keyed
  by the repository id, which is how a rename becomes a move instead of a
  second clone. A run holds a lock on it; a second fetch exits `2` while the
  first runs.

## 5. Read the results

- The human plan collapses routine fetches into its summary and prints rows
  only for drift; `--format=json` lists every action. Use JSON with jq when
  you need to report or filter.
- `orphan` rows are reports, never actions: a duplicate of a pinned checkout,
  a remembered checkout outside the layout, or a repository GitHub no longer
  returns by id (`remote-gone`). They are never fetched, moved or deleted.
- status rows that failed to inspect carry `error`, and the command exits `1`
  after printing the others.

| Code | fetch | status |
| --- | --- | --- |
| `0` | plan rendered, or apply without failures or conflicts | snapshot rendered, dirty and diverged rows included |
| `1` | transport, Git, state or output error | a checkout, discovery or state failure |
| `2` | usage error, bad config, missing token or owner, state locked | usage error or bad config |
| `3` | apply left a failure, a conflict or an unsaved state | — |

Fields and examples: [references/output.md](references/output.md).

## 6. The local service

`maintainer service run` serves maintainer over HTTP at
`http://maintainer.localhost:17841/`: `/status` is the status table in the
browser, `/inbox` lists every issue the checkouts mirror in `.github/issues`
with a reading pane, filters and a score, `/tablo` shows them on one board
(Backlog, To do, In progress, Done, Won't do); `/` redirects to `/status`.
It reads the fetch config and state as `status` does, from the home
directory, and `service.toml` in `$XDG_CONFIG_HOME/maintainer` or
`~/.config/maintainer` for the addresses (`listen`) and the score weights; on
the first start it writes that file with every key commented out, so read it
to learn where the service listens (`--config` names another file, which must
exist). It needs no token or network.

- **It writes issue files.** Setting `impact: *` or `effort: *` on a page
  changes those label lines in the issue file of the checkout and nothing
  else. Don't run `dump`, `sync` or push the mirror, and don't edit the issue
  files, while the user triages in the browser; after you changed issue files,
  **Reindex** in the navigation bar rebuilds the index.

- **Homebrew runs it.** On macOS the Formula carries the service:
  `brew services start octolab/tap/maintainer`, `restart` after an upgrade,
  `stop` to end it; it logs to `$(brew --prefix)/var/log/maintainer.log`.
  Installing the Formula doesn't start it; start or stop it only with the
  user's go-ahead.
- **As the user, never as root.** Don't use `sudo maintainer` or
  `sudo brew services`; nothing here needs an administrator.
- **Loopback only.** By default it listens on `127.0.0.1:17841` and
  `[::1]:17841`, as `maintainer.localhost` resolves to both; without an IPv6
  loopback it skips `::1` and keeps running.
- **`--listen` replaces both** and may repeat. Each value is a literal
  loopback IP address with a port; `localhost`, `0.0.0.0` or another address
  exits `2`, and an address it cannot bind, such as a taken port, stops it
  with an error:

  ```sh
  maintainer service run --listen 127.0.0.1:8080 --listen '[::1]:8080'
  ```

- **Check it** with
  `curl -fsS http://maintainer.localhost:17841/status -o /dev/null -w '%{http_code}\n'`,
  which prints `200`. Run in a terminal, `Ctrl-C` stops it once the
  responses in flight are sent.

## 7. When something fails

- `unknown key …; did you mean …?`: fix the key in the named file, then run
  `maintainer fetch config validate`.
- `a token is required`: the profile has no token; ask the user to export
  `GITHUB_TOKEN` or set `token_file`, never to paste the token to you.
- A `conflict` names its path and reason: an occupied or foreign target, an
  `incomplete clone`, a missing pin, two checkouts of one repository. Report
  it; the user decides, the rest of the plan still runs.
- Exit `3` after apply: read the per-repository errors on stderr and the
  plan's summary; whatever succeeded stays applied, and the next apply retries
  the rest.
- A rate limit makes fetch wait once; a wait longer than an hour or than
  `--timeout` stops it with the wait named.

Report what you ran, what came back, and anything left as a conflict or an
error. Don't claim a fetch succeeded from its exit code alone.
