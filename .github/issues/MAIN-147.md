---
code: MAIN-147
id: null
databaseId: null
number: null
url: null
title: "fetch: worktrees: keep linked worktrees usable and read them as Git does"
issueType: Bug
labels:
  - "type: bug"
  - "scope: code"
  - "scope: test"
  - "scope: docs"
  - "severity: minor"
  - "impact: medium"
  - "effort: medium"
assignees:
  - kamilsk
milestone: "[[MAIN-M3]]"
parent: null
issueFields: {}
state: OPEN
stateReason: null
duplicateOf: null
createdAt: null
updatedAt: null
lastEditedAt: null
closedAt: null
---

# fetch: worktrees: keep linked worktrees usable and read them as Git does

## Problem

A move must keep a checkout usable as a whole, but when fetch moves a checkout after a rename or a transfer on GitHub, its linked worktrees keep pointing at the old location: Git refuses to work in them until `git worktree repair` runs in the moved checkout. Neither the plan nor the apply mentions them, so the move quietly breaks a worktree the user works in. The other way round, a checkout that is itself a linked worktree has its origin read from another place than the one Git uses: fetch sees no origin where `git config --local` shows one, and an origin update written the way Git writes it ([[MAIN-135]]) would not be what fetch reads back.

## PoC

What a move does to a worktree, with plain Git:

```bash
set -o pipefail; export GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_NOSYSTEM=1
tmp=$(mktemp -d) && cd "$tmp"
git init -q -b main a && git -C a -c user.name=t -c user.email=t@t commit -q --allow-empty -m init
git -C a worktree add -q ../w -b feature
mv a b                                         # a fetch move is one rename of the checkout
git -C w status                                # fatal: gitfile does not point to a valid repository
git -C b worktree repair && git -C w status --short   # works again
```

The fetch case is a service test built like the lifecycle tests of `service_test.go`: a tracked checkout with one linked worktree outside the workspace, its snapshot renamed, one apply.

Reading a linked worktree, as an adapter test in `internal/service/git`: a repository `a` with `origin = https://github.com/acme/tool.git` and a linked worktree `w` of it. `git -C w config --local remote.origin.url` prints the URL; `Inspect(w)` today returns no origin, while `Inspect(a)` returns it.

## Expected behavior

After a successful move, every linked worktree of the moved checkout works as before, or the apply names each worktree it could not bring back, counts it as an error and exits `3`. The plan of a move names the linked worktrees it affects. Reading a checkout's origin and push URLs uses the configuration Git uses for that checkout — for a linked worktree, the one of its main repository — so fetch reads exactly what `git config --local` shows and what an origin update writes.

## Acceptance criteria

- [ ] In the service test, `git -C <worktree> status` succeeds after the move, and the worktree's branch, HEAD, index and files are byte-identical to before.
- [ ] A worktree that cannot be brought back — its directory gone or unwritable — is named in the apply output and counted as an error; the move itself stands.
- [ ] The plan of a move lists the linked worktrees of the checkout, in the human and the JSON output.
- [ ] In the adapter test, `Inspect(w)` returns the same origin and push URLs as `git -C w config --local --get-all remote.origin.url` and `… remote.origin.pushurl`, and the same as `Inspect(a)`.
- [ ] `fetch.md` says how a move treats linked worktrees.

## Out of scope

Linked worktrees as checkouts of their own: discovering, adopting or moving them beyond reading their configuration consistently; moves across filesystems. The identity side of moves is [[MAIN-150]], and the origin update itself is [[MAIN-135]].

## Sources

- [[20261019T105848Z#F2. Moving a checkout breaks its linked worktrees]]
