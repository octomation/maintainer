---
title: Fetch repositories
description: Reconcile local GitHub checkouts across owners with a plan-first, non-destructive workflow.
---

# Fetch repositories

> [!NOTE]
> `maintainer fetch` is not released yet. To try it, build `maintainer` from a source checkout:
>
> ```sh
> git clone https://github.com/octomation/maintainer.git
> cd maintainer && go install .
> ```

`maintainer fetch` discovers GitHub repositories across several owners and
reconciles a local checkout tree, in the spirit of `terraform plan` / `apply`.
A local state file (keyed by the stable numeric repo `id`) remembers what was
materialised, so a rename or transfer on GitHub is detected as a **move**, not
as a delete-and-reclone.

It is **plan-only by default**: `--apply` performs clone, fetch refs, move,
update remote and adopt actions. Existing branches and working files are not
merged or reset. Orphan checkouts are retained and reported; they are never
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
discovery for the root command; it still loads the default state file.

```toml
[workspace]
root         = "."        # checkout root; defaults to the current directory
path         = "{{.Visibility}}/{{.Owner}}/{{.Repo}}"
pins         = []         # existing checkout trees to keep in place

[defaults]
clone_url    = "ssh"      # "ssh" | "https"
concurrency  = 4
# state_file = "/path/to/state.json"   # default: $XDG_STATE_HOME/maintainer/fetch/state.json

[filters]                 # gate new clone/adopt decisions; tracked repos stay tracked
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
explicitly. Narrow a single run with `--owner` (repeatable), e.g. `--owner acme`
over a `["*"]` config processes only `acme`.

Owner/profile selection narrows remote discovery for new clones. Existing
checkouts explicitly included by the workspace or per-repo pins can still be
verified and managed even when their owner is outside that remote selection.

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
path component. Per-repo pins retain the full rendering context including `.Root`.
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
`~/.dotfiles`, without moving the folder. Missing or foreign pinned paths are
conflicts; fetch never clones into a pin or falls back to a duplicate.

Use numeric IDs so rules survive renames. Applied pins are also retained as
`pinned_path` in state, protecting name-based rules on later runs. Removing a
rule alone does not unpin: deliberately remove its state `pinned_path` too to
return to template-managed moves, provided the checkout belongs to the current
workspace layout. Removing a workspace tree pin instead suspends its remembered
checkouts; it does not enable moves. Changing a per-repo pin selects another verified
existing checkout. Adoption/remote update fetch refs on the next invocation.

Use [`maintainer status`](/status/) to inspect local branches and divergence.

## Profiles & tokens

A profile is a `(token, owners)` pair. Token resolution order (per profile):
`token_file` (must be ≤ `0600`) → `token_env` (default `GITHUB_TOKEN` for the
first profile in alphabetical order) → inline `token` (warns).
`--token` overrides token resolution for that first configured profile, or
supplies the token in single-run mode. When the same repo is visible from two
profiles, the broader-visibility snapshot wins (private > public); the winning
profile's credentials/transport are used and recorded.

- **HTTPS + PAT** — the supported path for private repos; the token is passed
  per-operation (never written into the remote URL).
- **SSH** — best-effort via a running `ssh-agent` with strict known-hosts;
  `maintainer` does not manage keys.

`--profile <name>` (repeatable) limits a run to a subset of profiles.

## Plan & apply

| Action          | Trigger                                                                 |
| --------------- | ----------------------------------------------------------------------- |
| `clone`         | on the API, no matching checkout, target path absent                    |
| `fetch`         | tracked & present — `git fetch --prune` (remote-tracking refs only)     |
| `move`          | rendered path differs from state (e.g. a rename) — same-volume rename   |
| `relocate`      | state path missing, the same `id` found at exactly one other location   |
| `update_remote` | `remote.origin.url` drifted from the canonical URL                      |
| `adopt`         | a clone on disk matches an API repo with no state record                |
| `orphan`        | inactive duplicate, remembered checkout outside scope, or cached/confirmed remote disappearance; report only |
| `noop`          | no executable change, including filtered or inaccessible repositories; a reason may be shown |
| `conflict`      | ambiguous identity, missing/foreign pin, occupied target or another condition blocking an action |

Apply order: `adopt`/`relocate` → `update_remote` → `move` → `clone` → `fetch`
(clone/fetch bounded by `--concurrency`). The human plan collapses routine
fetches into the summary and prints lines only for drift; `--format=json` lists
every action. Two actions resolving to the same target path are a conflict for
both, decided up front.

Each `--apply` invocation discovers and builds a fresh plan; it does not read
back a previously printed JSON plan. Conflicting actions are skipped, while
independent executable actions can still run. Apply exits with `3` if conflicts
remain or any repository action fails.

Orphan rows identify their local path; JSON also includes `orphan_reason` and,
for inactive copies, `active_path`. Reasons are `duplicate-pin`, `out-of-scope`,
and `remote-gone`. A missing listing alone is not proof of deletion: remote-gone
requires a 404 from the repository-ID check. Only apply saves that observation
for offline status; authentication and transient failures do not establish it.

Target handling is fail-closed. Any existing clone or move target must first be
identified as the expected repository; otherwise the plan reports a `conflict`
with its path. Apply repeats the existence check immediately before clone/move,
so a path created after plan review is never overwritten. Clone reserves an
absent path atomically and never recursively cleans that final path on failure;
an unexpected partial checkout is retained with an error for manual inspection.

Repository identity comes only from `remote.origin.url` (the fetch endpoint).
`remote.origin.pushurl` is orthogonal: push locks such as a deliberately invalid
push URL neither create an ambiguous checkout nor get removed by
`update_remote`.

An accessible upstream with no refs is a valid fetch no-op. It is retried on
every run, so the first ref created later is discovered automatically.
Authentication, authorization, not-found and transport failures remain errors.

## State file

A single JSON document at `$XDG_STATE_HOME/maintainer/fetch/state.json`
(fallback `~/.local/state/maintainer/fetch/state.json`), `0600`, advisory-locked
for the run. Plan mode may create the state directory and `.lock` file, but
does not save repository state or modify checkouts. `id` is the
primary key; everything else is a last-observed value. `state prune` only forgets
records whose path is already gone — it never deletes a clone.

The fetch group also provides these subcommands:

| Command | Behavior |
| --- | --- |
| `fetch config init` | Write a TOML template to `./fetch.toml`, or the explicit `--config` path; `--force` overwrites it |
| `fetch config validate` | Parse and structurally validate an existing config without resolving tokens or calling GitHub |
| `fetch state show` | Print state as JSON using the configured/default state location |
| `fetch state prune` | Save state after forgetting missing paths; runs immediately, without `--apply` |

State subcommands acquire the same state lock. Use a nonempty `--config` path
to select a specific config for them.

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
as `<root>/…` and collapses routine fetches into the summary.

## Flags & exit codes

| Flag                    | Default | Notes                                            |
| ----------------------- | ------- | ------------------------------------------------ |
| `--apply`               | off     | execute the plan (otherwise plan-only)           |
| `--config <path>`       | auto    | `--config=""` disables discovery                 |
| `--profile <name>…`     | all     | limit to profiles                                |
| `--owner <name>…`       | all     | limit to owners                                  |
| `--token <token>`      | resolved | token override for the first/default profile   |
| `--format human\|json`  | human   | plan output (logs always go to stderr)           |
| `--concurrency <n>`     | `0`     | positive parallel discovery/clone/fetch cap; `0` uses config (default `4`) |
| `--timeout <dur>`       | `0`     | wall-clock budget                                |
| `-v`, `--verbose`       | `0`     | repeat to increase verbosity (`-vv`, `-vvv`)     |
| `-q`, `--quiet`         | off     | suppress routine progress; mutually exclusive with verbose |
| `-h`, `--help`          |         | show command help                               |

| Code | Meaning |
| --- | --- |
| `0` | Plan rendered successfully, or apply completed without action failures/conflicts |
| `1` | Transport, Git, state or output error |
| `2` | User input error, including bad config, missing token or lock contention |
| `3` | Apply finished with at least one per-repo failure or unresolved conflict |

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
maintainer fetch --apply         # adopt existing checkouts; clone missing repositories
maintainer fetch --apply         # refresh refs of newly adopted checkouts
maintainer fetch state show | jq '.repos | length'
```

Adoption matches clones by `remote.origin.url` → stable `id` (following GitHub's
rename redirect), so existing clones are reconciled in place rather than
re-cloned. A re-run is idempotent.

## Supported scope

Fetch manages one workspace and GitHub repository clones. It does not perform
GitHub writes, push, checkout, merge, reset, automatic deletion or archival.
Submodules and LFS content are not fetched. Managed moves require a destination
on the same filesystem; external pins stay in place. For local branch and
working-tree inspection, use [status](/status/); for layout and discovery
rules, see [workspace](/workspace/).
