> # 👨‍🔧 maintainer
>
> `maintainer fetch` — reconcile local GitHub checkouts.

`maintainer fetch` discovers GitHub repositories across several owners and
reconciles a local checkout tree, in the spirit of `terraform plan` / `apply`.
A local state file (keyed by the stable numeric repo `id`) remembers what was
materialised, so a rename or transfer on GitHub is detected as a **move**, not
as a delete-and-reclone.

It is **plan-only by default** and **safe on disk**: `--apply` performs only
non-destructive actions (clone, fetch refs, move, update remote, adopt). A repo
that disappears from GitHub is reported as an `orphan` (404-confirmed) and the
local clone is **retained, never deleted**.

## Quickstart

```bash
maintainer fetch config init          # write ./fetch.toml (--force to overwrite)
maintainer fetch config validate      # parse + structurally check

maintainer fetch                      # PLAN only — no disk writes
maintainer fetch --apply              # clone / fetch / move / update-remote / adopt

maintainer fetch --format=json | jq   # machine-readable plan (lists every action)
maintainer fetch state show | jq      # dump the state file
maintainer fetch state prune          # forget records whose path is gone
```

## Configuration

Read from `fetch.{toml,yaml}` (format by extension). Lookup order (first match
wins; a missing file is not an error):

1. `--config <path>` (`--config=""` disables file discovery).
2. `$MAINTAINER_FETCH_CONFIG`.
3. `./fetch.toml`, then `./fetch.yaml`.
4. `$XDG_CONFIG_HOME/maintainer/fetch.{toml,yaml}` (fallback `~/.config/maintainer/`).

```toml
[defaults]
root         = "."        # checkout root; defaults to the current directory
path         = "{{.Visibility}}/{{.Owner}}/{{.Repo}}"
clone_url    = "ssh"      # "ssh" | "https"
concurrency  = 4
# state_file = "/path/to/state.json"   # default: $XDG_STATE_HOME/maintainer/fetch/state.json

[filters]                 # gate only NEW clone decisions; tracked repos stay tracked
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
# path  = "~/Code/special"      # absolute/~ allowed only for per-repo overrides

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

### Path templates

`path` is a Go `text/template` with: `.Root .Owner .Repo .Visibility`
`.DefaultBranch .IsFork .IsTemplate .IsArchived` (plus `lower`/`upper` funcs).
Precedence high→low: per-repo → per-owner → `defaults.path`. The rendered path
is absolute → used as-is; `~` → expanded from `$HOME`; otherwise joined with
`root`. `defaults.path` and per-owner templates must stay **within `root`**;
absolute/`~` are allowed only for per-repo overrides.

## Profiles & tokens

A profile is a `(token, owners)` pair. Token resolution order (per profile):
`token_file` (must be ≤ `0600`) → `token_env` (default `GITHUB_TOKEN` for the
first profile) → inline `token` (warns). When the same repo is visible from two
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
| `orphan`        | gone on GitHub (404-confirmed) — clone retained, reported, never removed |
| `noop`          | everything matches                                                      |

Apply order: `adopt`/`relocate` → `update_remote` → `move` → `clone` → `fetch`
(clone/fetch bounded by `--concurrency`). The human plan collapses routine
fetches into the summary and prints lines only for drift; `--format=json` lists
every action. Two actions resolving to the same target path are a conflict for
both, decided up front.

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
(fallback `~/.local/state/…`), `0600`, advisory-locked for the run. `id` is the
primary key; everything else is a last-observed value. `state prune` only forgets
records whose path is already gone — it never deletes a clone.

## Flags & exit codes

| Flag                    | Default | Notes                                            |
| ----------------------- | ------- | ------------------------------------------------ |
| `--apply`               | off     | execute the plan (otherwise plan-only)           |
| `--config <path>`       | auto    | `--config=""` disables discovery                 |
| `--profile <name>…`     | all     | limit to profiles                                |
| `--owner <name>…`       | all     | limit to owners                                  |
| `--format human\|json`  | human   | plan output (logs always go to stderr)           |
| `--concurrency <n>`     | config  | parallel discovery/clone cap                     |
| `--timeout <dur>`       | `0`     | wall-clock budget                                |
| `-v`/`-q`               | —       | verbosity / quiet (mutually exclusive)           |

- `0` clean (incl. "no drift") · `1` transport/Git/state error ·
  `2` user input error (bad config/flags, missing token, lock contention) ·
  `3` apply finished with at least one per-repo failure or unresolved conflict
  (the summary lists which).

## Onboarding an existing tree

`adopt` is the read-only equivalent of `terraform import`: it lets `fetch` start
without breaking an existing checkout. First run against a tree laid out as
`<root>/{public,private}/<owner>/<repo>`:

```bash
cat > ~/.config/maintainer/fetch.toml <<'EOF'
[defaults]
root      = "/Users/me/Development"
clone_url = "ssh"
[filters]
exclude_archived = true
exclude_forks    = true
[profiles.primary]
token_env      = "GITHUB_TOKEN"
include_owners = ["*"]
EOF

maintainer fetch                 # review: adopt=<existing>, clone=<missing>
maintainer fetch --apply         # adopt writes state only; clone fetches the rest
maintainer fetch state show | jq '.repos | length'
```

Adoption matches clones by `remote.origin.url` → stable `id` (following GitHub's
rename redirect), so existing clones are reconciled in place rather than
re-cloned. A re-run is idempotent.

## Non-goals (PoC)

No GitHub writes, no submodules/LFS, no issue/PR/wiki ingestion, no working-tree
mutation (refs only), no auto-delete/auto-archive, no auto-move outside `root`,
no GraphQL (REST only — a deferred experiment), no daemon. See the
[PoC implementation plan](../.github/notes/) for the full design.
