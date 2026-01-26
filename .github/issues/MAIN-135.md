---
code: MAIN-135
id: null
databaseId: null
number: null
url: null
title: "fetch: identity: preserve checkout identity through repository changes"
labels:
  - "type: bug"
  - "scope: code"
  - "scope: test"
  - "severity: major"
  - "impact: high"
  - "effort: medium"
milestone: "[[MAIN-M3]]"
state: OPEN
stateReason: null
createdAt: 2026-09-26T09:24:00Z
updatedAt: 2026-09-26T09:24:00Z
lastEditedAt: null
closedAt: null
---

# fetch: identity: preserve checkout identity through repository changes

**Motivation:** the central promise of `fetch` is that a rename or a transfer on GitHub becomes a move of the existing checkout, keyed by the stable repository id, and that the checkout keeps its local branches, its push lock and its identity. The real plan of 2026-09-26 contains exactly such a transfer, `octolab/cdn.octolab.org` → `octopot/cdn.octolab.org`.

**Requirements**

- A checkout at a path that a state record attributes to a different id is a conflict that names both ids, never an adoption. The recorded id may win instead when the stored `remote_url` equals the checkout's origin.
- A rename that changes only letter case is not a conflict when `os.SameFile` confirms the destination is the checkout itself and the destination is otherwise safe: the checkout is renamed through a temporary name, or only the remote is updated. After a rename the record carries the new spelling; when only the remote was updated, the record keeps the on-disk spelling and the planner treats paths that are equal ignoring case and confirmed by `os.SameFile` as settled, so the next plan proposes no move. A different checkout at the destination, a failed identity verification or an unsafe destination stays a conflict.
- `UpdateRemote` changes only `remote.origin.url` in the checkout's own local Git config, reads the raw value back and succeeds exactly when it equals the desired URL — an already-correct URL is success, not an error, and leaves the file untouched. Every line other than origin's `url` keeps its bytes: push URLs, so a push lock survives a transfer; other remotes with their `url` and `pushurl`; keys without a value, such as `[commit] gpgsign`, which Git reads as true; line and trailing comments; `insteadOf` rules. The stored value is the canonical URL, not its `insteadOf` rewrite, so a local rule that matches it keeps applying. A path without its own repository is an error, and `GIT_DIR`, `GIT_COMMON_DIR` or `GIT_CONFIG` inherited from a Git hook never redirect the change to another repository.
- A move commits its new path to state as soon as the rename succeeds, before the remote update; a failed remote update therefore leaves a record at the new path, and the next run retries `update_remote` without proposing the move again.
- After a successful fetch the record carries the current snapshot's owner, name, visibility, default branch and flags, with `last_seen` updated.
- Tests: name reuse after a rename; case-only rename on a case-insensitive filesystem, and an occupied destination on a case-sensitive one; remote-only convergence; `UpdateRemote` with a local `insteadOf` that matches origin, with `pushurl = no_push`, with the desired URL already present, and on the PoC config compared byte for byte; `UpdateRemote` on a path without its own repository and under hook variables; a failed remote update after a successful rename, then a rerun; metadata refresh after an ordinary fetch. The raw Git-config change is a separately reviewable commit.

**Evidence:** findings 9 (`internal/service/fetch/adopter.go:122`, `planner.go:288`), 19 (`internal/service/fetch/occupancy.go:29`) and 23 (`internal/service/git/sync.go:152`) of [the code review](../reviews/20260926T071417Z.md), and G4 of [the review of 2026-07-10](../reviews/20260710T095546Z.md) (`planner.go:318`, `applier.go:114`), reproduced on `932b9c8`. The owner keeps five checkouts under `git lock`, which sets exactly the push URL of finding 23; that justifies regression coverage, not a claim that they are corrupted today.

**PoC**

An adapter test in `internal/service/git` appends this to the local config of a fresh repository, saves the file as `config.before` and calls `UpdateRemote(path, "https://github.com/acme/new.git")`:

```ini
# a line comment
[commit]
	gpgsign
[remote "origin"]
	url = https://github.com/acme/old.git
	fetch = +refs/heads/*:refs/remotes/origin/*
	pushurl = no_push
[remote "upstream"]
	url = https://github.com/upstream/tool.git ; a trailing comment
	fetch = +refs/heads/*:refs/remotes/upstream/*
	pushurl = DISABLE
[url "git@github.com:"]
	insteadOf = https://github.com/
```

The call returns nil, and in the checkout:

```sh
diff config.before .git/config                  # only origin's url line: acme/old.git → acme/new.git
git config --local --get-all remote.origin.url  # https://github.com/acme/new.git
git remote get-url origin                       # git@github.com:acme/new.git
git config --get-all remote.origin.pushurl      # no_push
git config --get-all remote.upstream.url        # https://github.com/upstream/tool.git
git config --get-all remote.upstream.pushurl    # DISABLE
git config --type=bool commit.gpgsign           # true
```

A second call with the same URL returns nil and leaves the file byte for byte as it is. With `GIT_DIR`, `GIT_COMMON_DIR` and `GIT_CONFIG` pointing at another repository, the call still changes only this checkout; on a directory without its own `.git` inside another repository it fails and changes nothing.

The name-reuse and case-only cases are planner and adopter tests: state `{1: acme/old}`, snapshots `{1: acme/new, 2: acme/old}` with a checkout at `acme/old` must yield a conflict or a move, never `adopt id=2` plus `clone id=1`; a snapshot renamed `octocat` → `OctoCat` with the checkout at `octocat` on a case-insensitive filesystem must yield a move through a temporary name or a remote update, and a second plan must be `fetch` only.

**Out of scope:** configuring `upstream` remotes ([[MAIN-128]]); moves across filesystems; submodules.

Related: [[MAIN-92]], [[MAIN-123]], [[MAIN-133]], [[MAIN-145]], [[MAIN-136]].

<!-- 2026-10-01T12:25Z
Preparation for implementation, from the spec reviews of #466 (2026-10-01): reconfirmed on `main` — `applier.go:79-97` runs `Move`, then `UpdateRemote` (returning before any commit on error), then one `commit`, so a failed remote update leaves state at the old path and recovery goes through `relocate`. Separately, `planner.go:288` adopts a workspace checkout with no path cross-check: a clone of a fork's parent at the fork's path is adopted as the parent and moved on the next run. #466 now specifies that a path belonging to another known id is a conflict, not an adoption; this identity work is the natural place for the check.
-->
