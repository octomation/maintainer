---
code:
id:
databaseId:
number:
url:
title: "fetch: state: honor config selection and preserve uncertain records"
labels:
  - "type: bug"
  - "scope: code"
  - "scope: test"
  - "severity: major"
  - "impact: medium"
  - "effort: easy"
milestone: "[[milestone-3]]"
state: OPEN
stateReason:
createdAt: 2026-09-26T09:23:00Z
updatedAt: 2026-09-26T09:23:00Z
lastEditedAt:
closedAt:
---

# fetch: state: honor config selection and preserve uncertain records

**Motivation:** the state file is the memory of what `fetch` manages. Its administration commands must operate on the store the user named, a record may be forgotten only when its path is known to be gone, and a failed save must never pass for a successful apply.

**Requirements**

- `--config=""` means the same for `state show` and `state prune` as for `fetch`: no config discovery and the default state location; an omitted flag keeps discovery. Covered for the environment variable and the working-directory config.
- `prune` forgets only records whose path does not exist (`ENOENT`); every other error keeps the record, is reported by path and reason, and the command exits `1`.
- A failed `Store.Save` during or after apply is an error: reported with the affected repositories, counted, and reflected in the exit code (`3` when actions ran, `1` otherwise), never ignored. The promise is preservation and explicit diagnostics, not silent repair: existing checkouts and local work stay untouched, and the next run recovers a record automatically only where the identity of the checkout on disk can be established safely — the same origin and, when present, the same recorded id — through relocate or adopt without re-cloning; otherwise the checkout is a conflict naming what is known. Fault-injection tests cover a move whose state save fails, and a move whose save and remote update both fail, so that the moved checkout still carries an old origin while its recorded path is absent.

**Evidence:** finding 2 of [the gpt-6 review](../reviews/20260926T071223Z.md) (`internal/command/fetch/state.go:83` against `internal/command/fetch/root.go:135`), G3 of [the review of 2026-07-10](../reviews/20260710T095546Z.md) (`internal/command/fetch/state.go:54`), and the ignored save results at `internal/service/fetch/service.go:453` and `:472`.

**PoC**

```bash
set -o pipefail; tmp=$(mktemp -d) && export XDG_STATE_HOME="$tmp/xdg" MAINTAINER_FETCH_CONFIG="$tmp/fetch.toml"
printf '[defaults]\nstate_file = "%s/other.json"\n[profiles.primary]\ntoken_env = "GITHUB_TOKEN"\n' "$tmp" > "$tmp/fetch.toml"
printf '{"version":1,"profiles":{},"repos":[{"id":1,"owner_login":"a","name":"b","path":"%s/gone","remote_url":"https://github.com/a/b.git"}]}' "$tmp" > "$tmp/other.json"
chmod 600 "$tmp/other.json"
maintainer fetch --config="" state show | jq '.repos | length'   # 1 today; expected: 0, the default empty store
maintainer fetch --config="" state prune                          # "pruned 1 record(s)" today; expected: nothing to prune
```

```bash
mkdir -p "$tmp/locked/repo" && chmod 000 "$tmp/locked"           # run as an unprivileged user
printf '{"version":1,"profiles":{},"repos":[{"id":2,"owner_login":"a","name":"c","path":"%s/locked/repo","remote_url":"https://github.com/a/c.git"}]}' "$tmp" > "$tmp/other.json"
maintainer fetch --config "$tmp/fetch.toml" state prune; echo "exit=$?"
jq '.repos | length' "$tmp/other.json"                            # today: the record is pruned, exit 0; expected: 1 record kept, the path reported, exit 1
chmod 755 "$tmp/locked"
```

The save failures are service tests with a store whose `Save` fails on demand and a fake Git port: after a successful rename and a failed save, the run exits `3` naming the repository; a rerun against the same disk plans a `relocate` (same origin, same id) and no clone; with the remote update also failed, the rerun reports a conflict for that checkout rather than adopting it under a name that may since have been reused.

**Out of scope:** any other state migration; changing the state file format or location.

Related: [[issue-workspace-enforce-boundaries]], [[issue-fetch-honor-owners-profiles-filters]], [[issue-fetch-refresh-reconciled-checkouts]].

<!-- 2026-10-01T12:25Z
Preparation for implementation, from the spec reviews of #466 (2026-10-01): reconfirmed on `main` — intermediate save errors during apply are discarded at `internal/service/fetch/service.go:453` and `:472`. The fork-remote work (#466) adds a second stage after clone and adopt that runs only after the first stage is saved, so it depends on these errors being surfaced.
-->
