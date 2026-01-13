> # 👨‍🔧 maintainer
>
> `maintainer status` — inspect local repository work.

Status reads the same `fetch.toml` / `fetch.yaml` and state as
[`maintainer fetch`](fetch.md), scans the checkout root, and includes remembered
paths and explicitly pinned checkouts such as `~/.dotfiles`. It needs Git on
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

In a terminal, use arrows or `j/k` to highlight a row, Page Up/Down to scroll,
`g/G` or Home/End for first/last, and left/right or `h/l` to scroll horizontally.
The footer shows the selected path, HEAD commit and upstream. Exit with `q`,
Escape or Ctrl-C. The terminal is restored on exit and resize adjusts the
viewport. Enter is reserved for future detail navigation.

Redirected output automatically uses the plain table. `--format=tui` forces
interactive mode and requires terminal input and output. `--concurrency N`
overrides the config's inspection cap; `--timeout 30s` bounds the whole command.
The released macOS and Linux targets support the terminal view.

Per-repo paths select the active clone and suppress its stale duplicates.
Without a pin, duplicates remain separate rows; use JSON or the selected-path
footer to distinguish them. ID-only pins without state can only use their
local origin for display identity: status cannot verify a GitHub ID offline.
Fetch performs that verification. If origin contradicts known state, status
reports the mismatch and suggests fetch. Missing pins remain error rows.
Ignore rules are respected; archived/fork filters do not hide local work.

Config discovery follows fetch, including `MAINTAINER_FETCH_CONFIG`, the current
directory and XDG paths. `--config=""` disables configuration discovery, not
state loading. `--root` changes the scan root; remembered state paths and pins
are still included. No recursive submodule scan or API discovery is performed.

Exit codes: `0` for a successful snapshot (including dirty/diverged rows), `1`
for inspection/state failures, `2` for invalid configuration/options. Other
repositories are still rendered when an individual checkout fails. JSON is an
array, including `[]` for an empty collection, with `repository`, `path`,
`branch`, `default_branch`, `commit`, `upstream`, `added`, `deleted`,
`changed_files`, `untracked`, `binary`, `conflicts`, `ahead`, `behind`, `status`,
`pinned`, optional `id`, and optional `error` fields.
