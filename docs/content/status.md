---
title: Repository status
description: Inspect branches, uncommitted changes and divergence across local checkouts, offline.
---

# Repository status

> [!NOTE]
> `maintainer status` is not released yet. To try it, build `maintainer` from a source checkout:
>
> ```sh
> git clone https://github.com/octomation/maintainer.git
> cd maintainer && go install .
> ```

Status reads the same `fetch.toml` / `fetch.yaml` and state as
[`maintainer fetch`](/fetch/), using the same [workspace](/workspace/) layout
and pins. It includes eligible remembered paths and explicitly pinned checkouts
such as `~/.dotfiles`. It needs Git on
PATH, but no token or network. It never fetches, modifies the index/worktree,
or writes state. Ahead/behind counts describe **local upstream refs**; run
`maintainer fetch --apply` separately to refresh them.

## Usage

```bash
maintainer status                         # navigable table in a terminal
maintainer status --format=plain          # all rows, printable, no ANSI
maintainer status --format=json           # all fields and raw counts
maintainer status --root ~/Development
maintainer status --owner kamilsk --owner octomation
maintainer status --config /path/to/fetch.toml
maintainer status --config="" --root ./repos
```

The command accepts no positional arguments. With no configuration, it uses
the default workspace layout under the current directory; `--root` selects
another root. This still reads the default fetch state file.

## Table columns

```text
REPOSITORY             │ BRANCH  │ UNCOMMITTED       │ STATUS     │ LOCK │ PATH
kamilsk/dotfiles        │ main*   │ +39/-0 · 3f       │ synced     │      │ ~/.dotfiles
kamilsk/mindset         │ main*   │ +0/-0             │ behind 409 │      │ public/kamilsk/mindset
octomation/maintainer   │ fetcher │ +12/-1 · 2f · ?1  │ ahead 1    │ 🔒   │ public/octomation/maintainer
```

`Path` is the last column in both plain and terminal tables. Checkouts inside
the effective `workspace.root` use paths relative to that root, including
relative pins and per-owner layouts. External overrides and remembered paths
under the home directory use `~/…`; paths outside both root and home remain
absolute. A checkout at root is shown as `.`, and an external checkout at home
as `~`. `--root` also changes the display base. Paths describe the actual local
checkout, including missing pins and orphans, rather than a future fetch target.
JSON `path` and the selected-path footer retain the full path.

`*` marks the default branch from local `origin/HEAD`, falling back to cached
fetch metadata. `+/-` counts tracked lines from HEAD to the combined current
index/worktree, with rename detection disabled. Rename-only changes therefore
count removed/added lines and remain visible. `f` counts changed tracked files;
`?` counts untracked files, `!` counts conflicted files. Binary changes have
their own count. In an unborn repository the line count describes staged files
against the empty tree. Untracked file contents are not added to line counts.

`synced` means the current branch matches its upstream; it can still have
uncommitted work. A branch can be both ahead and behind. Detached HEAD,
unborn branches, missing upstream configuration and deleted upstream refs are
reported distinctly. Failures produce error rows, not zero/clean results.

`Lock` shows `🔒` when any remote has a push URL exactly equal to `no_push`,
the marker set by the dotfiles helper `git lock [remote]` (`origin` by default).
Otherwise the cell is empty. It includes other remotes and multiple push URLs;
the marker does not imply that every push destination is blocked. Custom push
URLs, server permissions and hooks are not checked. After `git unlock [remote]`,
run status again to refresh the indicator. Fetch and branch status are independent
of this marker.

## Interactive controls

The selected row and active column have separate highlights; the footer shows
the full path, HEAD commit and upstream. The table is a snapshot taken at startup;
run the command again to refresh it.
Use up/down or `j/k` for rows, Page Up/Down to scroll, `g/G` or Home/End for
first/last, and `h/l` to scroll horizontally. Resize adjusts the viewport.

### Sorting

Select a column with Tab/Shift+Tab or left/right. Press `s` or Enter to sort by
that column, replacing the whole sort chain. Shift+s (`S`) or Shift+Enter adds
the column after the existing keys. Header clicks work the same way; Shift+click
adds a key. Repeated activation cycles **ascending → descending → removed**.
In additive mode, toggling a key keeps its priority; removing and re-adding it
moves it to the end. `0` clears all sort keys, retaining the current filter.

Headers show direction and precedence (`Branch ↑1`, `Status ↓2`). Sorting covers
the entire filtered collection, including rows outside the viewport. Repository
and branch compare case-insensitively. Path compares the displayed path text
case-sensitively. Uncommitted compares `(added + deleted,
changed files, untracked, binary, conflicts)` numerically; status compares
`(ahead + behind, ahead, behind, status text)`. Lock sorts unmarked rows first
ascending and marked rows first descending. Numeric error rows come last in
both directions. Ties use repository/path for deterministic order. Clearing
sorting restores repository/path order, or fuzzy relevance while searching.

Shift+Enter requires a terminal that reports modified keys. `S` works in legacy
terminals too. Some terminals reserve Shift+mouse for native text selection;
use `S` if Shift+click is intercepted.

### Fuzzy filtering

Press `/` or click the search line and start typing. Matches update immediately;
characters may be non-contiguous and matching ignores case (`dtf` finds
`dotfiles`). Space-separated terms must all match, and can match different
fields: repository, branch, uncommitted/status text, lock icon (`🔒`), displayed
or full path, upstream, or commit.
With no explicit sort, results use fuzzy relevance; explicit keys take precedence.
The header shows matched/total counts, and an empty result is shown explicitly.

Enter or Escape leaves the input while keeping the filter. Ctrl+u clears the
input. Outside the input, Escape clears an active filter; otherwise it exits.
`q` and Ctrl-C exit from the table; Ctrl-C also exits from the input, where `q`
is ordinary text. Sorting preserves the selected checkout. If filtering hides
it, clearing the filter restores it unless a different row was selected.
All sort/filter state is in memory. Plain/JSON output retains the original
complete inventory; no repository or fetch state is changed.

Auto mode uses the interactive table when input and output are terminals and
`TERM` is set to a value other than `dumb`. Otherwise it uses the plain table.
`--format=tui` forces interactive mode and requires terminal input and output.

## Workspace, pins and orphan rows

Per-repo paths select the active clone for mutation, while stale duplicates
remain visible as `orphan`. Their own branch, changes and
divergence are retained; plain output and TUI details identify the active path.
Orphan is a management warning, not a replacement for Git status or a read error.
Without a specific per-repo pin, duplicates remain separate error rows; broad
workspace pins do not choose a winner. The Path column distinguishes these
checkouts; JSON and the selected-path footer also retain their full paths.
ID-only pins without state can only use their
local origin for display identity: status cannot verify a GitHub ID offline.
Fetch performs that verification. If origin contradicts known state, status
reports the mismatch and suggests fetch. Missing pins remain error rows.
Ignore rules are respected; archived/fork filters do not hide local work.

An orphan row retains its Git status and adds a management warning. JSON
`orphan_reason` distinguishes `duplicate-pin` (another checkout is active),
`out-of-scope` (a remembered checkout is outside current management rules), and
`remote-gone` (fetch cached a GitHub 404 by repository ID). The last case includes
the cached check time. Status never probes GitHub or infers remote deletion
from stale refs. Search for `orphan` or its reason to find these rows.

## Configuration

Status uses the same configuration as [fetch](/fetch/#configuration), in order:

1. `--config <path>`; `--config=""` disables discovery.
2. An existing file named by `$MAINTAINER_FETCH_CONFIG`.
3. `./fetch.toml`, then `./fetch.yaml`.
4. `$XDG_CONFIG_HOME/maintainer/fetch.toml`, then `fetch.yaml`, falling back to
   `~/.config/maintainer/`.

The state location comes from `defaults.state_file`, or
`$XDG_STATE_HOME/maintainer/fetch/state.json` (fallback
`~/.local/state/maintainer/fetch/state.json`). Missing state is an empty inventory
of remembered checkouts. Status does not acquire a state lock or create state
directories. `--config=""` still loads state from the default location.

`--root` changes the workspace root and rebases relative pins.
State paths outside the current layout are report-only `orphan` rows,
not recursive scan roots; legacy explicit
per-repo pins remain supported. The default layout excludes `research` and
other unrelated branches without exclusion lists. No recursive submodule scan
or API discovery is performed.

## Flags

| Flag | Default | Meaning |
| --- | --- | --- |
| `--config <path>` | auto | Config discovery override; empty disables discovery |
| `--root <path>` | config, otherwise `.` | Override workspace root for this run |
| `--format auto\|plain\|json\|tui` | `auto` | Select output mode |
| `--owner <name>` | all local owners | Filter rows by owner, case-insensitively; repeatable |
| `--concurrency <n>` | config, otherwise `4` | Parallel Git inspections; explicit value must be positive |
| `--timeout <duration>` | `0` | Whole-command time limit, including TUI; `0` is unlimited |
| `-h`, `--help` | | Show command help |

## JSON output and exit codes

JSON is an array, including `[]` for an empty collection, with `repository`,
absolute `path`, `branch`, `added`, `deleted`,
`changed_files`, `untracked`, `binary`, `conflicts`, `ahead`, `behind`, `status`,
`push_locked` (boolean), and `pinned`. `id`, `default_branch`, `commit`,
`upstream`, and `error` are omitted when unavailable or empty. Orphan rows also expose
`orphan_reason`, optional `active_path`, and `remote_checked_at` for cached
`remote-gone` observations saved by fetch apply. Text/TUI shows only `orphan`,
without bracketed reason codes; the reason remains in JSON and searchable.

For example, list changed checkouts with their full paths:

```bash
maintainer status --format=json | jq '.[] | select(.changed_files > 0 or .untracked > 0) | {repository, path}'
```

List checkouts marked by `git lock`:

```bash
maintainer status --owner octomation --format=json | jq -c '.[] | select(.push_locked) | {repository, path}'
# {"repository":"octomation/maintainer","path":"/Users/kamilsk/Development/public/octomation/maintainer"}
```

| Code | Meaning |
| --- | --- |
| `0` | Successful snapshot, including dirty, diverged and orphan rows |
| `1` | Git inspection, discovery, state or output failure |
| `2` | Invalid configuration or options, including TUI without terminal I/O |

Individual checkout errors remain in the output alongside successful rows,
then the command exits with `1`. A fatal configuration, state or discovery
failure can prevent the snapshot from being rendered.
