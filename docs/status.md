> # 👨‍🔧 maintainer
>
> `maintainer status` — inspect local repository work.

Status reads the same `fetch.toml` / `fetch.yaml` and state as
[`maintainer fetch`](fetch.md), using the same [workspace](workspace.md) layout
and pins. It includes eligible remembered paths and explicitly pinned checkouts
such as `~/.dotfiles`. It needs Git on
PATH, but no token or network. It never fetches, modifies the index/worktree,
or writes state. Ahead/behind counts describe **local upstream refs**; run
`maintainer fetch --apply` separately to refresh them.

```bash
maintainer status                         # navigable table in a terminal
maintainer status --format=plain          # all rows, printable, no ANSI
maintainer status --format=json           # all fields and raw counts
maintainer status --root ~/Development
maintainer status --owner kamilsk --owner octomation
maintainer status --config /path/to/fetch.toml
maintainer status --config="" --root ./repos
```

```text
REPOSITORY             │ BRANCH  │ UNCOMMITTED       │ STATUS
kamilsk/dotfiles        │ main*   │ +39/-0 · 3f       │ synced
kamilsk/mindset         │ main*   │ +0/-0             │ behind 409
octomation/maintainer   │ fetcher │ +12/-1 · 2f · ?1  │ ahead 1
```

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

The terminal screen uses [Bubble Tea](https://github.com/charmbracelet/bubbletea),
Lip Gloss styling and the Bubbles text input. The selected row and active column
have separate highlights; the footer shows the path, HEAD commit and upstream.
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
and branch compare case-insensitively. Uncommitted compares `(added + deleted,
changed files, untracked, binary, conflicts)` numerically; status compares
`(ahead + behind, ahead, behind, status text)`. Numeric error rows come last in
both directions. Ties use repository/path for deterministic order. Clearing
sorting restores repository/path order, or fuzzy relevance while searching.

Shift+Enter requires a terminal that reports modified keys. `S` works in legacy
terminals too. Some terminals reserve Shift+mouse for native text selection;
use `S` if Shift+click is intercepted.

### Fuzzy filtering

Press `/` or click the search line and start typing. Matches update immediately;
characters may be non-contiguous and matching ignores case (`dtf` finds
`dotfiles`). Space-separated terms must all match, and can match different
fields: repository, branch, uncommitted/status text, path, upstream, or commit.
With no explicit sort, results use fuzzy relevance; explicit keys take precedence.
The header shows matched/total counts, and an empty result is shown explicitly.

Enter or Escape leaves the input while keeping the filter. Ctrl+u clears the
input. Outside the input, Escape clears an active filter; otherwise it exits.
`q` and Ctrl-C exit from the table; Ctrl-C also exits from the input, where `q`
is ordinary text. Sorting preserves the selected checkout. If filtering hides
it, clearing the filter restores it unless a different row was selected.
All sort/filter state is in memory. Plain/JSON output retains the original
complete inventory; no repository or fetch state is changed.

Redirected output automatically uses the plain table. `--format=tui` forces
interactive mode and requires terminal input and output. `--concurrency N`
overrides the config's inspection cap; `--timeout 30s` bounds the whole command.
The released macOS and Linux targets support the terminal view.

Per-repo paths select the active clone for mutation, while stale duplicates
remain visible as `orphan [duplicate-pin]`. Their own branch, changes and
divergence are retained; plain output and TUI details identify the active path.
Orphan is a management warning, not a replacement for Git status or a read error.
Without a specific per-repo pin, duplicates remain separate error rows; broad
workspace pins do not choose a winner. Use JSON or the selected-path footer to
distinguish them. ID-only pins without state can only use their
local origin for display identity: status cannot verify a GitHub ID offline.
Fetch performs that verification. If origin contradicts known state, status
reports the mismatch and suggests fetch. Missing pins remain error rows.
Ignore rules are respected; archived/fork filters do not hide local work.

Config discovery follows fetch, including `MAINTAINER_FETCH_CONFIG`, the current
directory and XDG paths. `--config=""` disables configuration discovery, not
state loading. `--root` changes the workspace root and rebases relative pins.
State paths outside the current layout are report-only `orphan [out-of-scope]`,
not recursive scan roots; legacy explicit
per-repo pins remain supported. The default layout excludes `research` and
other unrelated branches without exclusion lists. No recursive submodule scan
or API discovery is performed.

Exit codes: `0` for a successful snapshot (including dirty/diverged rows), `1`
for inspection/state failures, `2` for invalid configuration/options. Other
repositories are still rendered when an individual checkout fails. JSON is an
array, including `[]` for an empty collection, with `repository`, `path`,
`branch`, `default_branch`, `commit`, `upstream`, `added`, `deleted`,
`changed_files`, `untracked`, `binary`, `conflicts`, `ahead`, `behind`, `status`,
`pinned`, optional `id`, and optional `error` fields. Orphan rows also expose
`orphan_reason`, optional `active_path`, and `remote_checked_at` for cached
`remote-gone` observations saved by fetch apply. Status never probes GitHub or
infers remote deletion from stale local refs. Search for `orphan` to find them.
