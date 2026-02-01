---
code: MAIN-158
id: I_kwDOE2M9Zc8AAAABYustSQ
databaseId: 5954547017
number: 514
url: https://github.com/octomation/maintainer/issues/514
title: "service: tablo: the issues of the collection as a kanban board"
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
createdAt: 2026-10-21T04:50:57Z
updatedAt: 2026-10-21T10:24:10Z
lastEditedAt: 2026-10-21T10:24:10Z
closedAt: 2026-10-21T09:28:20Z
---

# service: tablo: the issues of the collection as a kanban board

## Motivation

The boards of the owner's projects differ, but most work fits five columns: Backlog, To do, In progress, Done, Won't do. `/tablo`, on the runtime of [[MAIN-155]], shows the issues of the collection on that one basic board, filtered the way the inbox ([[MAIN-157]]) filters, with the column derived by the runtime from the issue state and the board sections of its repository. Mapping each project's own statuses is a later concern.

## Expected behavior

- Five columns in the order Backlog, To do, In progress, Done, Won't do, each with its count from `facets.columns`; the board scrolls horizontally on a narrow screen.
- A card shows the code, the title, the repository, the impact and effort chips and the milestone; a flagged task (ambiguous boards or a contradiction between state and section) carries a visible marker with the reason in its tooltip.
- The shared `TaskFilters` of [[MAIN-155]], with the same parameters as the inbox; the page requests `state=all` by default, so Done and Won't do are populated; a milestone filter is the primary way to look at one release.
- Each column loads 50 cards from `GET /api/v0/tasks?column=…&limit=50` and offers "Load more": by default the open columns are ordered by score, Done and Won't do by `updated`, and the ordering of the Display menu (`sort` in the URL) overrides it for every column; the column counts come from `facets.columns`, which ignores the column selector; the accumulated pages reset when the filters or the `generation` change.
- A click on a card opens the shared `TaskDetails` as the full issue page in the content area (Escape or the breadcrumb returns to the board) with the content of `GET /api/v0/tasks/{id}`, including the triage buttons, so a task can be triaged from the board; after a successful triage the affected columns are requested again.
- The URL carries the filters, so a reload and a link restore the board.

## Acceptance criteria

- [ ] On the owner's collection the board shows the issues of this repository in the columns of `.github/roadmap/MAIN-P1.md` for its listed tasks, closed tasks under Done or Won't do by their state reason, and unlisted open tasks under Backlog.
- [ ] The milestone filter `MAIN-M3` shows only its tasks across the columns and the counts add up to the milestone's total.
- [ ] Component tests cover the column layout, the per-column paging, the reset on a filter change and the flag marker against the mocked API of [[MAIN-155]]; `npm run check` and `npm test` pass in the page's branch, the full `make web-check` is the integration gate.
- [ ] The page's Go side adds nothing beyond [[MAIN-155]]; if it needs a change there, the change is made in the runtime with its tests.

## Out of scope

Drag and drop between columns; writing the board files; per-project column sets; swimlanes; work-in-progress limits.

Related: [[MAIN-155]], [[MAIN-157]], [[MAIN-M3]].
