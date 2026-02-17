# Output

Logs and progress go to stderr; stdout carries only the result. Paths in JSON
are absolute, except the `path` of a status row with the status
`unresolved pin`, which keeps the rule's path as written.

## fetch --format=json

One object per run:

| Field | Meaning |
| --- | --- |
| `plan_id`, `generated_at` | the plan's identity and time |
| `discoveries` | per profile: the listings it read and how many repositories they returned |
| `actions` | every action of the plan, routine fetches included |
| `summary` | counts of the actions, keyed by the kinds below |
| `errors` | in `summary`: the failures of an apply, a failed state save included; `0` in a plan |

An action has these fields, the empty ones left out. One repository can have
several actions, such as a `move` followed by its `fetch`.

| Field | Meaning |
| --- | --- |
| `kind`, `id` | the kind below and the GitHub repository id |
| `owner`, `name`, `path` | the repository and its checkout |
| `from_path`, `to_path` | the old and new directory of a move or a relocation |
| `from`, `to`, `worktrees` | the old and new `owner/name` of a move, and the worktrees it reconnects |
| `reason`, `flag` | why, and a flag such as `unverified` or `inaccessible` |
| `orphan_reason`, `active_path` | for an orphan |

| Kind | Meaning |
| --- | --- |
| `clone` | on GitHub, not on disk; the target path is free |
| `fetch` | tracked and present: refresh its remote-tracking refs and their tags |
| `move` | its path changed, as after a rename or transfer |
| `relocate` | its recorded path is gone and the same id is at exactly one other place |
| `update_remote` | `remote.origin.url` differs from the canonical URL |
| `adopt` | a checkout on disk matches a repository with no state record |
| `orphan` | a report: inactive duplicate, out of scope, or `remote-gone` |
| `noop` | nothing to do; `reason` says why, such as `outside the selection` or a filter |
| `conflict` | blocked: ambiguous identity, occupied or foreign target, missing pin, a pin whose repository GitHub does not confirm, `incomplete clone`, a scan obstruction |

Apply prints the same plan with the resulting summary; it is not a log of each
action. Per-repository errors go to stderr.

```sh
maintainer fetch --format=json | jq '.summary | {conflict, orphan, clone, move}'
```

## status --format=json

An array of rows, `[]` when there are none:

| Field | Meaning |
| --- | --- |
| `id` | the GitHub repository id, from the fetch state or a `[[repos]]` rule |
| `repository`, `path` | `owner/name` from the origin or the state, and the checkout's directory |
| `branch`, `default_branch`, `commit`, `upstream` | the current branch and what it tracks |
| `added`, `deleted`, `changed_files`, `untracked`, `binary`, `conflicts` | uncommitted work against HEAD |
| `ahead`, `behind`, `status` | divergence from the local upstream ref, and its text such as `synced` |
| `push_locked` | a remote has the push URL `no_push` |
| `pinned` | kept in place by a pin |
| `visibility` | `public`, `private` or `internal` as the fetch state recorded it |
| `orphan_reason`, `active_path`, `remote_checked_at` | for orphan rows |
| `error` | the inspection failed; counts are unknown, not zero |

Empty optional fields are left out.

```sh
maintainer status --format=json | jq -c '.[] | select(.error) | {path, error}'
```

## Exit codes

| Code | fetch and its subcommands | status |
| --- | --- | --- |
| `0` | plan rendered, or apply finished without failures or conflicts | snapshot rendered |
| `1` | transport, Git, state or output error; `state prune` with a path it cannot check | a checkout failed to inspect, or discovery, state or output failed |
| `2` | usage error, bad config, missing token or owner, the state locked by another run | usage error, bad config, `--format=tui` without a terminal |
| `3` | apply left a failed repository, an unsaved state or a conflict | — |

| Code | skill | service |
| --- | --- | --- |
| `0` | done; `status` found no problem | `run` stopped by `Ctrl-C` or `SIGTERM`; `open` opened or printed the address |
| `1` | `status` found a problem; `install` or `export` left a directory alone or could not write | `run` could not bind an address; `open` found no service there, or another program answers |
| `2` | usage error, an unknown `--agent` or `--skill`, `show --file` of no file | usage error, bad `service.toml` or fetch config, `--listen` not a loopback address |

`maintainer github contribution suggest` exits `2` without `GITHUB_TOKEN` or
for an anchor in the future, and `4` when no moment fits up to now. A usage
error, such as an unknown flag or subcommand, always exits `2` before any
request.
