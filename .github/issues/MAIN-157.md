---
code: MAIN-157
id: I_kwDOE2M9Zc8AAAABYusqtA
databaseId: 5954546356
number: 513
url: https://github.com/octomation/maintainer/issues/513
title: "service: inbox: the issues of the collection as a mail app"
issueType: Feature
labels:
  - "scope: code"
  - "scope: test"
  - "type: feature"
  - "impact: high"
  - "effort: medium"
assignees:
  - kamilsk
milestone: "[[MAIN-M2]]"
parent: null
issueFields: {}
state: OPEN
stateReason: null
duplicateOf: null
createdAt: 2026-10-21T04:50:53Z
updatedAt: 2026-10-21T08:24:38Z
lastEditedAt: 2026-10-21T08:24:38Z
closedAt: null
---

# service: inbox: the issues of the collection as a mail app

## Motivation

Every checkout mirrors its issues in `.github/issues`, and the owner reads them repository by repository. `/inbox`, on the runtime of [[MAIN-155]], shows all of them in one place the way a mail client does: a list to scan, a pane to read, filters to narrow, an order that puts the important work first, and the two labels the owner triages with, `impact: *` and `effort: *`, editable in place.

## Expected behavior

- Two panes: the list on the left, the reading pane on the right; on a narrow screen the pane opens over the list.
- A list row shows the code, the title, the repository and the milestone, the workflow status and the age since `updatedAt` (the full date in a tooltip); the priority, estimate, labels and score are in the reading pane; the selected row is highlighted; the list loads 100 rows from `GET /api/v0/tasks` and offers "Load more".
- The shared `TaskFilters` of [[MAIN-155]] on top sends the query parameters of `GET /api/v0/tasks`: owner, repository, state (open by default, closed, all), milestone, labels, type, assignee and a text search; the options and counts come from `facets`.
- The ordering in the Display menu: score (default), priority, updated, created.
- The grouping in the Display menu: none (default), status, priority, estimate, repository, owner, milestone or type; a group is a header with its count over its rows, the ordering applies inside a group, and `group` is in the URL. The inbox has no list/board switch: the board is `/tablo`.
- The reading pane is the shared `TaskDetails` of [[MAIN-155]] fed by `GET /api/v0/tasks/{id}`: the title, a link to the issue on GitHub when it has a URL, the repository, state, milestone, type, assignees, dates, the label chips, the rendered body, the comments with their times and links, the provenance path, the score with its components, and the triage buttons.
- After a successful triage the page requests the list again with the same parameters and keeps the selection; a response that arrives for an older request is discarded; a `412` tells that the file changed and offers to reload the task; a `409` explains duplicate mirrors.
- Rows are focusable: Enter opens the task, Escape returns the focus to the list.
- The URL carries the filters and the sort (`?repo=…&state=…&sort=…`), so a reload and a link restore the view; the selected task is not in the URL.

## Acceptance criteria

- [ ] On the owner's collection the inbox lists the open issues of every mirrored repository, filters and sorts answer under 100 ms, and the body of a long issue with comments renders readably in both themes.
- [ ] Setting `impact: high` on an issue changes only that label line in the file (checked with `git diff` in the repository), the row and the pane reflect it, and a second edit after an external change of the file ends in `412` with the reload offer.
- [ ] Component tests cover the list, the sort select, the requery after a triage and the conflict flow against the mocked API of [[MAIN-155]]; `npm run check` and `npm test` pass in the page's branch, the full `make web-check` is the integration gate.
- [ ] The page's Go side adds nothing beyond [[MAIN-155]]; if it needs a change there, the change is made in the runtime with its tests.

## Out of scope

Editing the title, body or other labels; creating issues; pushing to GitHub (the mirror tooling does); read and unread state; notifications.

Related: [[MAIN-155]], [[MAIN-158]], [[MAIN-M2]].
