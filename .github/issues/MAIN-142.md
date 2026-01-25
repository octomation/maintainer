---
code: MAIN-142
id: null
databaseId: null
number: null
url: null
title: "status: inspection: recompute counters when a row is inspected again"
labels:
  - "type: bug"
  - "scope: code"
  - "scope: test"
  - "severity: minor"
  - "impact: low"
  - "effort: easy"
milestone: "[[MAIN-M3]]"
state: OPEN
stateReason: null
createdAt: 2026-09-28T07:30:00Z
updatedAt: 2026-09-28T07:30:00Z
lastEditedAt: null
closedAt: null
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
