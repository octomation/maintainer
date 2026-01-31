---
code: MAIN-143
id: null
databaseId: null
number: null
url: null
title: "status: overview: summarize the inventory and preserve the selection"
labels:
  - "type: improvement"
  - "scope: code"
  - "scope: test"
  - "scope: docs"
  - "impact: medium"
  - "effort: easy"
milestone: "[[MAIN-M3]]"
state: CLOSED
stateReason: COMPLETED
createdAt: 2026-09-26T09:30:00Z
updatedAt: 2026-09-26T09:30:00Z
lastEditedAt: null
closedAt: null
---

# status: overview: summarize the inventory and preserve the selection

**Motivation:** an overview of 160 checkouts starts with the totals — how many carry uncommitted work, how many are ahead or behind, locked, orphaned or failed — and the table must not lose the selected row under the user's hands.

**Requirements**

- The plain and TUI headers carry a summary computed from the existing rows: checkouts, with uncommitted changes, ahead, behind, locked, orphan, errors. Counts are rows and the categories overlap; a failed inspection is never counted as clean; in the TUI the summary says whether it describes the shown rows or the whole collection. JSON stays the existing array.
- When the filter hides the anchor, sorting keeps the highlighted row, and the original anchor is restored once the filter is cleared unless another row was selected. The selection fix and the summary are separate commits.
- `status.md` documents the summary and its meaning.

**Evidence:** the plain header at `internal/service/status/view.go:125` counts only checkouts; finding 28 of [the code review](../reviews/20260926T071417Z.md) (`internal/service/status/terminal.go:78`) reproduced the moving highlight with a model test.

**PoC**

```bash
maintainer status --format=plain | head -1
# today:    Repository status · 160 checkouts · local refs (no fetch)
# expected: Repository status · 160 checkouts · 81 uncommitted · 4 ahead · 0 behind · 5 locked · 4 orphan · 0 errors · local refs (no fetch)
```

The selection case is a model test: rows `alpha`, `beta1`, `beta2`, `gamma`; `G` selects `gamma`, a filter `beta` highlights `beta1`, `s` keeps `beta1`, and a second `s` must still highlight `beta1`, not `beta2`; clearing the filter must highlight `gamma` again.

**Out of scope:** a dashboard, configurable statistics, a JSON envelope, live refresh.

Related: [[MAIN-141]], [[MAIN-132]].
