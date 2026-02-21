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
  Requires the maintainer binary >=0.6.2 <0.7.0 on macOS or Linux and Git on
  PATH. fetch and the contribution commands need network access and a GitHub
  token; status works offline.
metadata:
  author: octomation
  version: "0.6.2"
  tool: maintainer
  tool-version-range: ">=0.6.2 <0.7.0"
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

This skill describes maintainer `>=0.6.2 <0.7.0`. Before anything else:

1. Run `maintainer version` and read the `version` line.
2. Compare it with that range.
   - In range: continue, and read `maintainer <command> --help` before using a
     command's flags.
   - Older or newer: don't rely on the recipes here. Read the skill that ships
     with the binary, `maintainer skill show`, and tell the user the versions
     differ. Run `maintainer skill status` to see where this skill comes from
     and suggest updating the side that is behind:
     - the binary: `brew upgrade --formula octolab/tap/maintainer`;
     - a `plugin: maintainer@octolab` copy, from the OctoLab catalog: the
       agent's plugin manager, `claude plugin update maintainer@octolab` in
       Claude Code, or `codex plugin marketplace upgrade octolab`, then
       `codex plugin add maintainer@octolab` in Codex;
     - a copy `installed by maintainer`: replace it from the binary,
       `maintainer skill install --agent claude-code --replace`, or `codex`.

     Never install a copy from the binary next to a plugin: the agent loads
     both.
   - `dev`, a pseudo-version such as `v0.6.2-0.20261126121749-fbe59efd127f`,
     sometimes ending in `+dirty`, or unknown: treat compatibility as unknown
     and rely on `--help`.
   - `maintainer skill` missing: the binary predates it; rely on `--help` only.
   - `maintainer` not found: explain how to install it
     (https://maintainer.octolab.org/quickstart/); don't install it without
     the user's go-ahead.
3. Never guess a flag or a config key that `--help` or the configuration
   guide, https://maintainer.octolab.org/fetch/#configuration, doesn't show.

For a review report of a repository, use the review skill that comes with
this one: `maintainer:review` in the plugin, or `review` from the binary.

## 2. Ground rules

- **Plan first; apply only with the user's go-ahead.** `maintainer fetch`
  without `--apply` changes no checkout and saves no repository state. Show
  the user the plan, in particular `conflict`, `move`, `relocate`,
  `update_remote` and `orphan` rows, and run `maintainer fetch --apply` only
  when they agree. Each `--apply` builds a fresh plan; it never replays a
  plan printed earlier.
- **Tokens stay secret.** Never print `GITHUB_TOKEN`, a `token_file` or an
  inline `token`, and never put a token on the command line: `--token` is
  removed and exits `2`. Ask the user to set the token themselves: for
  `fetch`, in its profile (see §4); for `github contribution`, which reads
  nothing else, as `GITHUB_TOKEN` in the environment.
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
| A first config | `fetch config init` (writes `~/.config/maintainer/fetch.toml`, or a `.toml` `--config` path only; `--force` overwrites) |
| To check a config | `fetch config validate` |
| To see what fetch would do | `fetch`, or `fetch --format=json` |
| To clone missing and refresh existing checkouts | `fetch --apply`, after reviewing the plan |
| One owner or one account only | `fetch --owner acme` or `fetch --profile bot` |
| Work in progress across checkouts | `status --format=plain` or `status --format=json` |
| Checkouts of some owners | `status --owner acme --format=json` |
| What fetch remembers | `fetch state show` |
| To forget checkouts whose directory is gone | `fetch state prune` (writes the state at once; `--owner` and `--profile` narrow it) |
| A time to contribute | `github contribution suggest`; also `contribution lookup`, `contribution histogram`, `contribution snapshot`, `contribution diff` |
| maintainer in a browser on this computer | `service run`, or `brew services start octolab/tap/maintainer`, then `service open` |

The subcommands of `fetch` list its flags under Global Flags but use only
`--config`, and `state prune` also `--owner` and `--profile`; `--format` is
ignored: `config validate` prints `ok: <path> (…)`, `state show` always prints
JSON. Recipes:
[references/workflows.md](references/workflows.md). Every command and flag of
this release: [references/cli.md](references/cli.md).

## 4. Configuration, state and tokens

- **Lookup**, first match wins: `--config <path>` (`--config=""` disables
  discovery), the file named by `$MAINTAINER_FETCH_CONFIG`, then
  `$XDG_CONFIG_HOME/maintainer/fetch.toml` or `.yaml` (usually
  `~/.config/maintainer/`). A `fetch.toml` in the working directory is never
  read: fetch and status print a notice, or exit `2` when there is no other
  config; pass it with `--config` only if the user says it is theirs. Without
  a config, `fetch --config=""` needs an `--owner` and works under the current
  directory.
- **The schema is the contract.** The keys are `[workspace]` (`root`, `path`,
  `pins`), `[defaults]` (`clone_url`, `concurrency`, `state_file`),
  `[filters]`, `[profiles.<name>]`, `[[owners]]` and `[[repos]]`; every key
  with its default is in the guide,
  https://maintainer.octolab.org/fetch/#configuration. An unknown key or
  table is bad config: fetch, its subcommands and status exit `2` and suggest
  the key it was probably meant to be. Fix the key; don't remove what you
  don't understand.
- **Layout is scope.** `workspace.path`, by default
  `{{.Visibility}}/{{.Owner}}/{{.Repo}}`, is both where new clones go and
  the only place both commands look. Checkouts elsewhere are managed only
  through `workspace.pins` or a per-repository `path` in `[[repos]]`; match
  `[[repos]]` by `id` so a rule survives a rename.
- **Rules the user may need**, in their config:
  - `[[repos]]` with `ignore = true` drops a repository from fetch and status:
    the fix for a confirmed `orphan`, or, matched by `owner` and `name`, for a
    checkout behind SSO that the token cannot read.
  - `clone_url = "https"` (default `"ssh"`) in `[defaults]`, a profile,
    `[[owners]]` or `[[repos]]`, the most specific winning: HTTPS with the
    profile's token is the supported way to clone a private repository, SSH
    needs the user's `ssh-agent`.
  - `[[owners]]` with `name` and `path` gives one owner its own layout
    template, relative to `workspace.root`.
- **Profiles and tokens.** A profile is a token with the owners it discovers
  (`include_owners`, `["*"]` for the user and every organization they belong
  to, `["*", "other"]` adds an owner outside them). Its token comes from `token_file` (mode `0600` at most), then
  `token_env` (`GITHUB_TOKEN` for the first profile in alphabetical order),
  then an inline `token`. fetch needs a token even for public repositories;
  status needs none. The `github contribution` commands ignore the profiles
  and take the token only from `GITHUB_TOKEN`.
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

- **Which checkouts.** `[fetch] config` in `service.toml` names the fetch
  config, relative to the home directory: set it when the config is not in
  `~/.config/maintainer`, since `brew services` starts the service without
  the shell's `$MAINTAINER_FETCH_CONFIG`. Without any fetch config it starts
  and reads the default layout under the home directory; a `fetch.toml` in
  the home directory is never read, and alone it stops the start with exit
  `2`. `[[repos]]` with `match` and `weight` (default `1`) multiplies the
  score of one repository's issues.

- **It writes issue files.** Setting `impact: *` or `effort: *` on a page
  changes those label lines in the issue file of the checkout and nothing
  else. Don't run `dump`, `sync` or push the mirror, and don't edit the issue
  files, while the user triages in the browser; after you changed issue files,
  **Reindex** in the navigation bar rebuilds the index.

- **Open it with `maintainer service open`.** Each start serves the data
  only at a new private address, `http://<label>.maintainer.localhost:<port>/`;
  the command checks the running service and opens that address in the
  browser, `--print` prints it instead. Run it after every start. The plain
  `http://maintainer.localhost:17841/` loads the pages without data and every
  `/api/` request there but `hello` gets `401`. The label is a key: don't
  print it into shared text, logs or files. With `access = "open"` in
  `service.toml` the plain address serves the data too: suggest it only to a
  user who is the only account of the computer.

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
- **`--listen` replaces both** and may repeat; for each port the service also
  holds the other of `127.0.0.1` and `::1`. Each value is a literal
  loopback IP address with a port; `localhost`, `0.0.0.0` or another address
  exits `2`, and an address it cannot bind, such as a taken port, stops it
  with an error that names it:

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
- `a token is required` or `no token resolved` from `fetch`: the profile has
  no token; ask the user to set its `token_file` or export its `token_env`
  (`GITHUB_TOKEN` for the first profile). `a GitHub token is required` from
  `github contribution` (exit `2`): only `GITHUB_TOKEN` in the environment
  helps, a `token_file` does not. Never ask for the token itself.
- A `conflict` names its path and reason: an occupied or foreign target, an
  `incomplete clone`, a missing pin, a pin whose repository GitHub does not
  confirm, two checkouts of one repository. Report
  it; the user decides, the rest of the plan still runs.
- Exit `3` after apply: read the per-repository errors on stderr and the
  plan's summary; whatever succeeded stays applied, and the next apply retries
  the rest.
- To see more of a `fetch` run, add `-vv` (discovery, confirmations and
  state saves) or `-vvv` (each Git operation with its duration and error);
  the lines go to stderr.
- `service open` says the service is not running when it was started on
  another port, such as with `service run --listen`: name that address,
  `maintainer service open --listen 127.0.0.1:8080`. The log of a service
  under `brew services` is in `$(brew --prefix)/var/log/maintainer.log`.
- A rate limit makes fetch wait once; a wait longer than an hour or than
  `--timeout` stops it with the wait named.

Report what you ran, what came back, and anything left as a conflict or an
error. Don't claim a fetch succeeded from its exit code alone.
