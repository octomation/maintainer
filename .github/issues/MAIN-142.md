---
code: MAIN-142
id: I_kwDOE2M9Zc8AAAABYurxBg
databaseId: 5954531590
number: 500
url: https://github.com/octomation/maintainer/issues/500
title: "status: inspection: recompute counters when a row is inspected again"
issueType: null
labels:
  - "scope: code"
  - "scope: test"
  - "type: bug"
  - "severity: minor"
  - "impact: low"
  - "effort: easy"
assignees: []
milestone: "[[MAIN-M3]]"
parent: null
issueFields: {}
state: CLOSED
stateReason: COMPLETED
duplicateOf: null
createdAt: 2026-10-21T04:49:19Z
updatedAt: 2026-10-21T04:51:26Z
lastEditedAt: 2026-10-21T04:51:26Z
closedAt: 2026-10-21T04:49:22Z
---

# status: inspection: recompute counters when a row is inspected again

**Motivation:** `Inspect` takes a `Row` and returns it enriched, but it adds to the counters it finds instead of setting them. A row that has already been inspected doubles its changes on the next pass. No caller does that today — `Collect` builds a fresh row for every checkout (`internal/service/status/service.go:128-143`) — so the defect is latent. It becomes visible as soon as anything refreshes an existing row: a TUI reload, a watch mode, or a retry after a timeout.

**Requirements**

- `Inspect` resets every value it derives from Git before reading it: `Changed`, `Untracked`, `Conflicts`, `Added`, `Deleted`, `Binary`, `Ahead`, `Behind`, `Commit`, `Branch`, `Upstream`, `Status`, `Error`, `PushLocked`. Metadata supplied by the caller (`Path`, `ID`, `Repository` and `DefaultBranch` from state, pins, orphan fields, `displayPath`) is kept.
- A test inspects the same row twice and gets identical results, including for a row that previously ended in an error.

**Evidence:** the counters are incremented in place at `internal/service/status/git.go:109-116` and `git.go:185-186`; found during the consultation on the CI failure of `TestInspectRenameAndConflict` ([[MAIN-127]]), reproduced on `932b9c8`.

**PoC**

```go
path := checkout(t, t.TempDir())                  // helper from git_test.go
write(t, filepath.Join(path, "text"), "one\nthree\n")
write(t, filepath.Join(path, "new"), "x\n")
first := Inspect(context.Background(), Row{Path: path})
second := Inspect(context.Background(), first)
// today:    first f=1 ?=1 +1/-1, second f=2 ?=2 +2/-2
// expected: second equals first
```

**Out of scope:** adding a refresh to the TUI; the rename-detection mode of `git status`.

Related: [[MAIN-141]].
