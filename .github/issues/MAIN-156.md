---
code: MAIN-156
id: I_kwDOE2M9Zc8AAAABYusoYA
databaseId: 5954545760
number: 512
url: https://github.com/octomation/maintainer/issues/512
title: "service: status: the checkout table in the browser"
issueType: Feature
labels:
  - "scope: code"
  - "scope: test"
  - "type: feature"
  - "impact: medium"
  - "effort: medium"
assignees:
  - kamilsk
milestone: "[[MAIN-M3]]"
parent: null
issueFields: {}
state: CLOSED
stateReason: COMPLETED
duplicateOf: null
createdAt: 2026-10-21T04:50:49Z
updatedAt: 2026-10-21T10:24:04Z
lastEditedAt: 2026-10-21T10:24:04Z
closedAt: 2026-10-21T09:28:19Z
---

# service: status: the checkout table in the browser

## Motivation

`maintainer status` shows the collection in a terminal table with sorting, fuzzy search and a summary line. The same view belongs in the hub at `/status`, on the runtime of [[MAIN-155]]: a GUI instead of the TUI, with the same rows and the same order, plus what a browser adds, a properties panel and a refresh that does not block the page.

## Expected behavior

- The page renders the cached snapshot from `GET /api/v0/checkouts` as a table with the TUI columns: Repository, Branch (the default branch marked with `*`), Uncommitted (`+a/-d · nf · ?u · b binary · !c`), Status, Lock, Path; the cells are the TUI cells, produced by the server.
- The rows are grouped by state (errors, uncommitted, ahead or behind, orphans, clean), by owner or not at all, chosen in the Display menu; a list/board switch shows the same groups as board columns (`layout=board` in the URL).
- The Display menu carries a summary line that follows the TUI header over the rows shown: checkouts, uncommitted, ahead, behind, locked, orphan, errors, and "local refs (no fetch)".
- The ordering in the Display menu: a click on a column cycles ascending, descending, off; a shift-click appends the column to the sort chain, as `Shift+s` does; the chain is shown there and sent as `sort`; the order, the fuzzy relevance of an unsorted search, unknown numbers last and the repository and path tie-break come from the exported status adapter on the server, never from the browser.
- The search of the Filter menu sends `q`; the results update without a new Git collection.
- A click on a row or a card opens a properties panel with the full path, upstream, commit, pinned and orphan information, the active path of a duplicate, the remote check time and the error text.
- Refresh in the Display menu posts `/api/v0/checkouts/refresh`, shows the collection in progress and the time of the snapshot shown; the first visit shows "Reading local repository status…" until the first snapshot exists.
- Error rows render as the TUI does: `?` for the counts and `error: …` in Status.
- A row lays its properties out in columns of fixed width, aligned from row to row, so values of different length do not shift the next column.
- Rows are focusable: Enter opens the panel of the focused row, Escape closes it.

## Acceptance criteria

- [ ] For the fixtures of `internal/service/status`, the page's requests to `GET /api/v0/checkouts` with a sort chain and a query render the cells, the order and the summary that the exported adapter gives `Plain` for the same input (component test against the JSON fixtures of [[MAIN-155]]).
- [ ] On the owner's collection the table shows every row of `maintainer status --format=plain`, sorting and searching answer under 100 ms, and refresh completes without freezing the page.
- [ ] `npm run check` and `npm test` pass in the page's branch; the full `make web-check`, with the rebuilt `web/dist`, is the integration gate; the page has no dependency beyond the shell's.

## Out of scope

Actions on checkouts (fetch, sync, lock); columns the TUI does not have; a live update without refresh.

Related: [[MAIN-155]], [[MAIN-143]], [[MAIN-M3]].
