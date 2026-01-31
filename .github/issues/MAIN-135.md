---
code: MAIN-135
id: I_kwDOE2M9Zc8AAAABYd-JBA
databaseId: 5937006852
number: 487
url: https://github.com/octomation/maintainer/issues/487
title: "fetch: remote: change only the origin URL when updating a remote"
issueType: null
labels:
  - "scope: code"
  - "scope: test"
  - "type: bug"
  - "severity: major"
  - "impact: high"
  - "effort: medium"
assignees:
  - kamilsk
milestone: "[[MAIN-M3]]"
parent: null
issueFields: {}
state: CLOSED
stateReason: COMPLETED
duplicateOf: null
createdAt: 2026-10-20T06:24:47Z
updatedAt: 2026-10-20T12:38:04Z
lastEditedAt: 2026-10-20T06:26:02Z
closedAt: 2026-10-20T06:24:50Z
---

# fetch: remote: change only the origin URL when updating a remote

**Motivation:** a rename or a transfer on GitHub ends with an origin update of the existing checkout: after a move, in a remote-only reconciliation, and in the adoption of a pinned checkout. That update must change exactly origin's fetch URL and nothing else that the user keeps in the checkout's Git configuration — push locks, other remotes and their disabled push, settings such as commit signing, comments, `insteadOf` rules — because a later release cannot restore what an earlier one rewrote. It is the one behaviour fix in the code of v0.2.0; the identity work around moves is [[MAIN-150]].

**Requirements**

- `UpdateRemote` changes only `remote.origin.url` in the checkout's own local Git config, reads the raw value back and succeeds exactly when it equals the desired URL — an already-correct URL is success, not an error, and leaves the file untouched. Every line other than origin's `url` keeps its bytes: push URLs, so a push lock survives a transfer; other remotes with their `url` and `pushurl`; keys without a value, such as `[commit] gpgsign`, which Git reads as true; line and trailing comments; `insteadOf` rules. The stored value is the canonical URL, not its `insteadOf` rewrite, so a local rule that matches it keeps applying. A path without its own repository is an error, and `GIT_DIR`, `GIT_COMMON_DIR` or `GIT_CONFIG` inherited from a Git hook never redirect the change to another repository.
- Tests: `UpdateRemote` with a local `insteadOf` that matches origin, with `pushurl = no_push`, with the desired URL already present, and on the PoC config compared byte for byte; `UpdateRemote` on a path without its own repository and under hook variables. The raw Git-config change is a separately reviewable commit.

**Evidence:** finding 23 (`internal/service/git/sync.go:152`) of [the code review](../reviews/20260926T071417Z.md), reproduced on `932b9c8`, and the effects on other remotes, bare booleans and comments in [the issue review of 2026-10-19](../reviews/20261019T105848Z.md#main-135-checkout-identity). The owner keeps five checkouts under `git lock`, which sets exactly the push URL of finding 23; that justifies regression coverage, not a claim that they are corrupted today.

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

**Out of scope:** the identity of checkouts through renames, transfers and name reuse ([[MAIN-150]]); configuring `upstream` remotes ([[MAIN-128]]); reading the configuration of a checkout that is a linked worktree ([[MAIN-147]]); moves across filesystems; submodules.

Related: [[MAIN-150]], [[MAIN-129]], [[MAIN-147]], [[MAIN-92]], [[MAIN-123]], [[MAIN-133]], [[MAIN-145]], [[MAIN-136]].
