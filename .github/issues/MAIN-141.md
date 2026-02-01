---
code: MAIN-141
id: I_kwDOE2M9Zc8AAAABYursLQ
databaseId: 5954530349
number: 499
url: https://github.com/octomation/maintainer/issues/499
title: "status: inventory: preserve distinct checkouts and useful errors"
issueType: null
labels:
  - "scope: code"
  - "scope: test"
  - "type: bug"
  - "severity: major"
  - "impact: high"
  - "effort: easy"
assignees: []
milestone: "[[MAIN-M2]]"
parent: null
issueFields: {}
state: CLOSED
stateReason: COMPLETED
duplicateOf: null
createdAt: 2026-10-21T04:49:11Z
updatedAt: 2026-10-21T10:23:20Z
lastEditedAt: 2026-10-21T04:51:23Z
closedAt: 2026-10-21T04:49:14Z
---

# status: inventory: preserve distinct checkouts and useful errors

**Motivation:** the overview is only useful when an error row means an error and an orphan row means a management warning. Two unrelated local-only projects with the same directory name must not become "duplicate" errors, an orphan row must keep its Git status, and a broken repository must show Git's own message rather than a blame on `git config`.

**Requirements**

- Duplicates are counted only among rows whose `owner/name` identity comes from origin or state; rows with an `OrphanReason` are excluded from the count and never overwritten, and keep their own branch, changes and divergence.
- Checkouts without a GitHub identity — no `origin`, another remote name, a local or `file://` URL — are ordinary rows keyed by path, never duplicates of each other.
- An error row carries the trimmed stderr of the failing Git command and names the step that failed. `git status` runs first; the lock probe runs independently afterwards, so the primary error is preserved and the lock marker still survives a later inspection error, as §9 of the status specification promises.

**Evidence:** findings 13 (`internal/service/status/service.go:161`, `:73`), 14 (`internal/service/status/service.go:47`, `git.go:89`) and 27 (`internal/service/status/git.go:59`, `git.go:84`) of [the code review](../reviews/20260926T071417Z.md), reproduced on `932b9c8`.

**PoC**

```bash
set -o pipefail; export XDG_STATE_HOME=$(mktemp -d) GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_NOSYSTEM=1
tmp=$(mktemp -d) && mkdir -p "$tmp/public/alice/tool" "$tmp/public/bob/tool"
git -C "$tmp/public/alice/tool" init -q && git -C "$tmp/public/bob/tool" init -q
maintainer status --config="" --root "$tmp" --format=json | jq -c '.[] | {repository, status, error}'
# today: two rows with status "error" and "multiple checkouts for this repository; select one with a per-repo path", exit 1
# expected: two rows with status "unborn" and no error, exit 0
mkdir -p "$tmp/public/x/broken" && printf 'gitdir: /nonexistent/x\n' > "$tmp/public/x/broken/.git"
maintainer status --config="" --root "$tmp" --format=json | jq -r '.[] | select(.error) | .error'
# today: git config: exit status 128 — expected: a message naming git status and quoting "fatal: not a git repository"
```

The orphan case is a service test: a state record at `A/public/acme/tool` and a second clone with the same origin under `B`; `status --root B` must return the `B` row with its Git status and the `A` row as `out-of-scope`, both without an error, exit `0`.

**Out of scope:** the summary line and the TUI selection ([[MAIN-143]]); resolving a symlinked working directory (finding 26, deferred with the documented `--root` workaround).

Related: [[MAIN-145]], [[MAIN-144]].
