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
| `summary` | counts by kind: `clone`, `fetch`, `move`, `relocate`, `update_remote`, `adopt`, `orphan`, `noop`, `conflict`, plus `errors` after an apply |

An action has `kind`, `id` (the GitHub repository id) and, as they apply,
`owner`, `name`, `path`, `from_path` and `to_path` for moves and relocations,
`from` and `to` for the old and new `owner/name` of a move, `worktrees` that a
move reconnects, `reason`, `flag` (such as `unverified` or `inaccessible`),
and for an orphan `orphan_reason` and `active_path`. One
repository can have several actions, such as a `move` followed by its `fetch`.

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
| `conflict` | blocked: ambiguous identity, occupied or foreign target, missing pin, `incomplete clone`, a scan obstruction |

Apply prints the same plan with the resulting summary; it is not a log of each
action. Per-repository errors go to stderr.

```sh
maintainer fetch --format=json | jq '.summary | {conflict, orphan, clone, move}'
```

## status --format=json

An array of rows, `[]` when there are none:

| Field | Meaning |
| --- | --- |
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

`maintainer github contribution suggest` exits `4` when no moment fits up to
now. A usage error, such as an unknown flag or subcommand, always exits `2`
before any request.
