---
title: Fetch repositories
description: Reconcile local GitHub checkouts across owners with a plan-first, non-destructive workflow.
---

# Fetch repositories

`maintainer fetch` discovers GitHub repositories across several owners and
reconciles a local checkout tree, in the spirit of `terraform plan` / `apply`.
A local state file (keyed by the stable numeric repo `id`) remembers what was
materialised, so a rename or transfer on GitHub is detected as a **move**, not
as a delete-and-reclone. If a new repository takes the old name before fetch
sees the rename, the checkout stays with the repository it was recorded for:
it moves with the rename while its recorded remote is still its origin, and
the new repository is cloned at the old path by the apply after that.
Otherwise the checkout is a conflict that names both repository ids. Fetch
never adopts it as the new repository or clones the renamed one again.

It is **plan-only by default**: `--apply` performs clone, fetch refs, move,
update remote and adopt actions. Existing checkouts and local work are
retained: a fetch updates and prunes remote-tracking refs and adds the tags
that point into them, while branches and working files are not merged or
reset. Orphan checkouts are retained and reported; they are never
automatically deleted, fetched or moved.

## Quickstart

```bash
maintainer fetch config init          # write ./fetch.toml (--force to overwrite)
maintainer fetch config validate      # parse + structurally check

maintainer fetch                      # plan only — no checkout or state changes
maintainer fetch --apply              # clone / fetch / move / update-remote / adopt

maintainer fetch --format=json | jq   # machine-readable plan (lists every action)
maintainer fetch --profile primary --owner acme
maintainer status                     # inspect local work after fetching
maintainer fetch state show | jq      # dump the state file
maintainer fetch state prune          # forget records whose path is gone
```

Edit the generated config to set `workspace.root`, owners and token sources
before running fetch. Fetch requires network access and a token, including
public-only discovery; [status](/status/) works offline without credentials.
The root command accepts flags and no positional arguments.

Without a config file, provide an owner and set `GITHUB_TOKEN` (or `--token`):

```bash
maintainer fetch --config="" --owner acme       # plan under the current directory
```

This uses the default layout and state location. Set `workspace.root` in a
config to choose another root; fetch has no `--root` flag.

## Configuration

Read from `fetch.{toml,yaml}` (format by extension). Lookup order (first match
wins; no discovered file selects the single-run mode above):

1. `--config <path>` (`--config=""` disables file discovery).
2. An existing file named by `$MAINTAINER_FETCH_CONFIG`.
3. `./fetch.toml`, then `./fetch.yaml`.
4. `$XDG_CONFIG_HOME/maintainer/fetch.{toml,yaml}` (fallback `~/.config/maintainer/`).

An explicit missing `--config` file is an error. `--config=""` disables config
discovery for fetch and its subcommands; they still use the default state file.

The keys, their types and defaults are defined by the schema
[`internal/config/fetch.cue`](https://github.com/octomation/maintainer/blob/main/internal/config/fetch.cue);
`cue vet -c fetch.cue fetch.toml` checks a config against it. A key or table
outside the schema is bad config: `fetch`, its subcommands and `status` stop
with exit `2` and name each one, its table and the file, with the key it was
probably meant to be:

```text
$ maintainer fetch config validate
Error: fetch config "/Users/me/fetch.toml": unknown key include_owner in [profiles.primary]; did you mean include_owners?
```

```toml
[workspace]
root         = "."        # checkout root; defaults to the current directory
path         = "{{.Visibility}}/{{.Owner}}/{{.Repo}}"
pins         = []         # existing checkout trees to keep in place

[defaults]
clone_url    = "ssh"      # "ssh" | "https"
concurrency  = 4
# state_file = "/path/to/state.json"   # default: $XDG_STATE_HOME/maintainer/fetch/state.json

[filters]                 # skip new clones and adoptions; tracked repos stay tracked
exclude_archived  = false
exclude_forks     = false
exclude_templates = false

[profiles.primary]
token_env      = "GITHUB_TOKEN"
include_owners = ["*"]     # see "Owner selection" below

[profiles.bot]            # a second account (e.g. a bot) with its own token
token_env      = "BOT_TOKEN"
include_owners = ["acme-bot"]
clone_url      = "https"   # HTTPS + PAT is the supported path for private repos

# [[owners]]               # per-owner path/clone_url override
# name = "acme"
# path = "{{.Visibility}}/{{.Owner}}/{{.Repo}}"

# [[repos]]                # per-repo override (matched by id or owner/name)
# match = { id = 12345678 }     # id-match survives a rename
# path  = "~/Code/special"      # select an existing checkout outside the layout

# [[repos]]
# match  = { id = 99999999 }    # silence a confirmed orphan
# ignore = true
```

### Owner selection (`include_owners`)

`include_owners` does double duty: it picks **which REST endpoint** is called
for each owner (you → `/user/repos`; a member org → `/orgs/{org}/repos?type=all`;
anyone else → public only) **and** acts as an allowlist applied after discovery.

| Value                         | Meaning                                                        |
| ----------------------------- | -------------------------------------------------------------- |
| `["acme", "acme-bot"]`        | exactly these owners                                           |
| `["*"]` or omitted            | **you + every org you are a member of** (from `/user/orgs`)    |

The wildcard expands from your org memberships, so you never enumerate orgs and
ones you join later are picked up automatically. It does **not** include repos
you only collaborate on in orgs you are *not* a member of — list those
explicitly.

Owners are compared case-insensitively, as GitHub compares them: `["Acme"]`
and `["acme"]` select the same repositories, and so do the owners of
`[[owners]]` and `[[repos]]` rules. `include_owners` selects only the discovery
of new repositories: a full run still verifies and manages tracked checkouts
and pins of other owners.

### Narrowing a run

`--owner` and `--profile` (both repeatable) narrow the whole invocation:
`--owner acme` over a `["*"]` config lists only `acme`'s repositories and
works on them alone. A repository outside the selection gets no request by
id or by name, no clone, fetch, move or remote update, and no new state. JSON
lists a tracked one as a `noop` with the reason `outside the selection`; the
human output collapses those rows and shows the selection in the summary,
for example `selected=12/40`.

| Tracked repository | In a narrowed run |
| --- | --- |
| Of a selected owner, transferred to another owner | Confirmed by id, then left alone: a `noop`, `transferred outside the selection to <owner>/<name>` |
| Of another owner, listed now under a selected owner | Moved and its origin updated, like any transfer |
| Of another profile, listed by a selected profile | Fetched with the selected profile's credentials; that profile is recorded after a successful apply |
| Of another owner or profile, not listed by a selected profile | Left alone: a `noop`, `outside the selection` |

A per-repository pin is explicit intent: an `id`-only pin that no selected
listing returns is confirmed once by id and managed only when its owner is
selected. Untracked checkouts of other owners are neither resolved nor
adopted.

### Path templates

The [workspace](/workspace/) is shared with status. `workspace.path` is both
the target layout and the discovery boundary: the default only scans
`{public,private,internal}/<owner>/<repo>`, not arbitrary root descendants.
Legacy `defaults.root/path` remain supported, but cannot be mixed with
`[workspace]`. `workspace.pins` explicitly includes additional existing checkout
trees without permitting moves. Adding a new root child does not expand scope.

Managed templates support scalar `.Owner .Repo .Visibility .DefaultBranch
.IsFork .IsTemplate .IsArchived`, literals, and `lower`/`upper`; conditionals and
other non-reversible expressions are rejected. Each scalar stays within one
path component. Per-repo paths can also use `.Root`. A per-repo path is
rendered only when every field it references is known. Fetch takes the fields
from GitHub or from the state record it keeps for the repository; a field that
neither supplies makes the pin unresolved, a conflict for that repository.
Offline, [status](/status/) knows `.Root`, the `.Owner` and `.Repo` of a rule
that matches by owner and name, and the fields of the state record; for any
other field it shows an `unresolved pin` row. A path that cannot be rendered
for one repository, such as a layout with `.DefaultBranch` and a branch named
`release/v1`, is a conflict for that repository only; the rest of the plan
goes on. So is a per-repo path that is not a valid template, such as
`~/ext/{{.Owner`; `fetch config validate` rejects it, naming the rule.
Precedence high→low: per-repo → per-owner → `workspace.path`. The rendered path
is absolute → used as-is; `~` → expanded from `$HOME`; otherwise joined with
`root`. `workspace.path` and per-owner templates must stay **within `root`**;
absolute/`~` paths are supported for per-repo overrides and workspace pins.
Relative `workspace.root` is resolved from the current working directory,
not the config directory. Relative pin paths are resolved from that root.

### Active checkouts outside the tree

Every per-repo `path` pins an **existing** active checkout. For example:

```toml
[[repos]]
match = { id = 154873464 } # kamilsk/dotfiles
path = "~/.dotfiles"
```

Fetch verifies its GitHub identity and selects it even if an old duplicate
exists at `Development/public/kamilsk/dotfiles`. Apply adopts the selected path
into state; the duplicate stays untouched and appears as a report-only
`orphan` with the active path (`orphan_reason: duplicate-pin` in JSON). Rename/transfer updates origin at
`~/.dotfiles`, without moving the folder. Adoption only adds the record and
keeps the checkout's Git configuration as it is: when the origin no longer
matches the repository's current name, the next apply updates it as a separate
`update_remote`. Missing or foreign pinned paths are conflicts; fetch never
clones into a pin or falls back to a duplicate.

Use numeric IDs so rules survive renames. Applied pins are also retained as
`pinned_path` in state, protecting name-based rules on later runs. Removing a
rule alone does not unpin: deliberately remove its state `pinned_path` too to
return to template-managed moves, provided the checkout belongs to the current
workspace layout. Removing a workspace tree pin instead suspends its remembered
checkouts; it does not enable moves. The current rule takes precedence over the
saved `pinned_path` whenever it can be rendered, so changing a per-repo pin
selects another verified existing checkout; the previous path stays in
`previous_paths` and as an `orphan` row. The apply that adopts a pin or updates
its origin also fetches its refs. A checkout adopted at the old path of a
renamed repository is moved, and fetched again, by the next apply.

Use [`maintainer status`](/status/) to inspect local branches and divergence.

## Profiles & tokens

A profile is a `(token, owners)` pair. Token resolution order (per profile):
`token_file` (must be ≤ `0600`) → `token_env` (default `GITHUB_TOKEN` for the
first profile in alphabetical order) → inline `token` (warns).
`--token` overrides token resolution for the first selected profile in that
order, or supplies the token in single-run mode. When the same repo is visible from two
profiles, the broader-visibility snapshot wins (private > public); the winning
profile's credentials/transport are used and recorded.

- **HTTPS + PAT** — the supported path for private repos; the token is passed
  per-operation (never written into the remote URL).
- **SSH** — best-effort via a running `ssh-agent` with strict known-hosts;
  `maintainer` does not manage keys.

`--profile <name>` (repeatable) narrows the run to these profiles; see
[Narrowing a run](#narrowing-a-run). A tracked repository missing from the
listings is checked by id with its own profile's token, and only that token
establishes a 404 as `remote-gone`. When its profile is not available, for
example after it was renamed in the config, the check uses the first selected
profile: a 404 or 403 then leaves the record unverified, a `noop` flagged
`unverified` that names both profiles, and nothing is saved.

A checkout of a repository that the token may not read, such as one in an
organization behind SAML single sign-on or an IP allowlist, or one blocked for
legal reasons, becomes a `conflict` row with GitHub's answer, and the rest of
the plan proceeds. An `ignore` rule matched by owner and name keeps fetch from
looking such a checkout up at all.

When GitHub's rate limit is exhausted, fetch waits once, until the limit
resets or for as long as GitHub asks; a secondary limit that does not say how
long waits at least a minute. Then fetch tries again once. A wait longer than
an hour, or longer than what is left of `--timeout`, stops the run at once,
and the error names the wait. Server errors (`502`, `503`, `504`) and dropped
connections are retried a few times; a refused request, an unknown host, a
TLS failure or any other error that does not pass on its own is not.

## Plan & apply

| Action          | Trigger                                                                 |
| --------------- | ----------------------------------------------------------------------- |
| `clone`         | on the API, no matching checkout, target path absent                    |
| `fetch`         | tracked & present — `git fetch --prune` (remote-tracking refs and their tags) |
| `move`          | rendered path differs from state (e.g. a rename) — same-volume rename   |
| `relocate`      | state path missing, the same `id` found at exactly one other location   |
| `update_remote` | `remote.origin.url` drifted from the canonical URL                      |
| `adopt`         | a clone on disk matches an API repo with no state record, unless filters exclude it |
| `orphan`        | inactive duplicate, remembered checkout outside scope, or a repository not accessible by id; report only |
| `noop`          | no executable change, including filtered, unverified or inaccessible repositories and those outside the selection; a reason may be shown |
| `conflict`      | ambiguous identity, missing/foreign pin, occupied or unsafe target, scan obstruction or another condition blocking an action |

Apply order: `adopt`/`relocate` → `update_remote` → `move` → `clone` → `fetch`
(clone/fetch bounded by `--concurrency`). The human plan collapses routine
fetches into the summary and prints lines only for drift; `--format=json` lists
every action. Two actions resolving to the same target path, or to targets one
inside the other, are a conflict for both, decided up front; so are a target
inside an existing checkout of the workspace and the action on that checkout.

Every `adopt`, `relocate`, `update_remote` and `move` of an existing checkout
is followed by a `fetch` of that checkout in the same apply. The plan counts
it, the human plan marks it with `+ fetch remote-tracking refs`, and JSON
lists it as a separate `fetch` action. It runs with the other fetches, last,
and only if the action before it succeeded. If the action succeeds and the
fetch fails, the action stays applied, the error is counted, and the next
apply fetches again. A checkout whose origin names a repository that took its
old name is not fetched through that origin: it is fetched after the next
apply updates the origin.

Each `--apply` invocation discovers and builds a fresh plan; it does not read
back a previously printed JSON plan. Conflicting actions are skipped, while
independent executable actions can still run. Apply exits with `3` if conflicts
remain or any repository action fails.

A state save that fails during or after apply is an error as well: it names
the repositories whose change it leaves unrecorded, counts in `errors`, and
apply exits with `3`, or `1` when no action ran. Their checkouts and local work
stay as applied. The next run records such a checkout again only when its
origin and id match the record, for example as a `relocate` of a moved
checkout; a checkout whose origin is still recorded for another repository is
a `conflict` naming that record.

Orphan rows identify their local path; JSON also includes `orphan_reason` and,
for inactive copies, `active_path`. Reasons are `duplicate-pin`, `out-of-scope`,
and `remote-gone`. `remote-gone` keeps its name for compatibility and means
*not accessible by id*: the repository-ID check with the record's own profile
returned 404, because the repository was deleted or access to it was lost. A missing listing alone does
not establish it. Only apply saves that observation, with its check time, for
offline status; authentication and transient failures do not establish it.

Target handling is fail-closed. Any existing clone or move target must first be
identified as the expected repository; otherwise the plan reports a `conflict`
with its path. A target whose parent inside the root is a symlink or a file is
a conflict too. Apply repeats these checks immediately before each clone or
move, so a path or a link created after plan review is refused; changes that
another process makes while a clone or move runs are not guarded against. Clone reserves an
absent path atomically and never recursively cleans that final path on failure.

A clone that stops before it finishes, cut by `--timeout`, a signal or a
dropped connection, keeps what it has already written for inspection, marked as
incomplete. Every later plan reports that path as a `conflict` with the reason
`incomplete clone` and neither adopts nor fetches it; remove the directory to
clone the repository again. A clone that fails before it receives any data
leaves nothing behind, and the next apply clones again. A clone of an empty
repository is complete.

On a case-insensitive filesystem, as on macOS by default, a rename that changes
only letter case keeps the directory as it is spelled and updates only its
origin. On a case-sensitive filesystem it is an ordinary move, and a different
checkout at the new spelling is a conflict.

If the origin update of a move fails, the checkout stays at its new path, and
the next apply retries only the update.

A move takes the checkout's linked worktrees along. The plan lists them under
the move (`+ reconnect linked worktree` in the human plan, `worktrees` in
JSON), and the apply reconnects each one to the moved checkout, as
`git worktree repair` does; a worktree inside the checkout moves with it.
A worktree that cannot be reconnected, because its directory is gone, its
`.git` file is not writable or links to another repository, or Git's record of
it cannot be read, is named as an error and the apply exits with `3`; the move
itself stands. Once the cause is fixed, run
`git worktree repair` in the moved checkout.

Repository identity comes only from `remote.origin.url` (the fetch endpoint).
For a checkout that is itself a linked worktree, it is read from the
configuration of its main repository, as `git config --local` reads it.
`remote.origin.pushurl` is orthogonal: push locks such as a deliberately invalid
push URL neither create an ambiguous checkout nor get removed by
`update_remote`. It changes only `remote.origin.url` through Git, so it needs
`git` on PATH; the rest of `.git/config` stays as written.

An accessible upstream with no refs is a valid fetch no-op. It is retried on
every run, so the first ref created later is discovered automatically.
Authentication, authorization, not-found and transport failures remain errors.

## State file

A single JSON document at `$XDG_STATE_HOME/maintainer/fetch/state.json`
(fallback `~/.local/state/maintainer/fetch/state.json`), `0600`, advisory-locked
for the run. Plan mode may create the state directory and `.lock` file, but
does not save repository state or modify checkouts. `id` is the
primary key; everything else is a last-observed value. `state prune` forgets
only records whose path does not exist. A path it cannot check, such as one
under a directory it cannot read, keeps its record and is reported with the
reason, and the command exits `1`. The checkouts themselves are retained.

The fetch group also provides these subcommands:

| Command | Behavior |
| --- | --- |
| `fetch config init` | Write a TOML template to `./fetch.toml`, or the explicit `--config` path; `--force` overwrites it |
| `fetch config validate` | Parse and structurally validate an existing config without resolving tokens or calling GitHub |
| `fetch state show` | Print state as JSON using the configured/default state location |
| `fetch state prune` | Save state after forgetting missing paths; runs immediately, without `--apply` |

State subcommands acquire the same state lock and select the config as fetch
does: `--config=""` disables config discovery and uses the default state
location.

## JSON output

`--format=json` writes an object with `plan_id`, `generated_at`, `discoveries`,
`actions` and `summary`. Action entries contain `kind`, `id`, and applicable
identity/path/reason fields. Moves and relocations use `from_path` and `to_path`;
orphan rows can include `orphan_reason` and `active_path`. Paths are absolute.
The summary counts actions by kind, plus apply errors; one repository can have
multiple actions, including an orphan row for an inactive copy.

```bash
maintainer fetch --format=json | jq '.actions[]? | select(.kind == "conflict" or .kind == "orphan")'
```

Apply emits the plan with the resulting summary after execution. Progress and
per-repository error messages go to stderr. JSON is a plan and summary, not a
per-action execution log. Human output abbreviates paths under the workspace
as `<root>/…`, prints other paths in full, including a sibling of the root such
as `~/Development-old`, and collapses routine fetches into the summary.

## Flags & exit codes

| Flag                    | Default | Notes                                            |
| ----------------------- | ------- | ------------------------------------------------ |
| `--apply`               | off     | execute the plan (otherwise plan-only)           |
| `--config <path>`       | auto    | `--config=""` disables discovery                 |
| `--profile <name>…`     | all     | limit this run to these profiles                 |
| `--owner <name>…`       | all     | limit this run to these owners                   |
| `--token <token>`      | resolved | token for the first selected profile            |
| `--format human\|json`  | human   | plan output (logs always go to stderr)           |
| `--concurrency <n>`     | `0`     | positive parallel discovery/clone/fetch cap; `0` uses config (default `4`) |
| `--timeout <dur>`       | `0`     | wall-clock budget; `0` is unlimited              |
| `-v`, `--verbose`       | `0`     | repeat to increase verbosity (`-vv`, `-vvv`)     |
| `-q`, `--quiet`         | off     | suppress routine progress; mutually exclusive with verbose |
| `-h`, `--help`          |         | show command help                               |

| Code | Meaning |
| --- | --- |
| `0` | Plan rendered successfully, or apply completed without action failures/conflicts |
| `1` | Transport, Git, state or output error |
| `2` | Usage error, bad config, missing token or owner, or lock contention |
| `3` | Apply finished with at least one per-repo failure, failed state save or unresolved conflict |

A usage error stops the command before any request, names its cause in
stderr and exits `2`: an unknown subcommand, such as `fetch state shwo`, an
unknown flag, a flag value that does not parse, such as `--timeout=soon`, a
negative `--concurrency` or `--timeout`, an unknown `--format`, or an
unexpected argument.

A successfully rendered plan exits `0` even when it contains conflicts. For
automation, inspect `summary.conflict` as well as the exit code.

## Onboarding an existing tree

`adopt` records an existing checkout in state during apply. First run against
a tree laid out as `<root>/{public,private,internal}/<owner>/<repo>`:

```bash
cat > fetch.toml <<'EOF'
[workspace]
root      = "/Users/me/Development"
pins      = ["prototyping"] # omit if this directory does not exist

[defaults]
clone_url = "ssh"
[filters]
exclude_archived = true
exclude_forks    = true
[profiles.primary]
token_env      = "GITHUB_TOKEN"
include_owners = ["*"]
EOF

maintainer fetch                 # review: adopt=<existing>, clone=<missing>
maintainer fetch --apply         # adopt and fetch existing checkouts; clone missing repositories
maintainer fetch --apply         # moves checkouts adopted at an old path; otherwise fetches only
maintainer fetch state show | jq '.repos | length'
```

Filters skip new clones and adoptions: an untracked fork or archived checkout
on disk is left as it is, in the layout, at a per-repository pin or under a
workspace pin alike, and JSON lists it as a `noop`. A repository that is
already tracked stays tracked.

Adoption matches clones by `remote.origin.url` → stable `id` (following GitHub's
rename redirect), so existing clones are reconciled in place rather than
re-cloned. A checkout found at the old path of a renamed repository is moved,
and fetched, by the second apply. From then on, a re-run against unchanged
repositories only fetches.

## Supported scope

Fetch manages one workspace and GitHub repository clones. It does not perform
GitHub writes, push, checkout, merge, reset, automatic deletion or archival.
Submodules and LFS content are not fetched. Managed moves require a destination
on the same filesystem; external pins stay in place. For local branch and
working-tree inspection, use [status](/status/); for layout and discovery
rules, see [workspace](/workspace/).
