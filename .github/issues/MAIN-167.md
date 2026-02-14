---
code: MAIN-167
id: null
databaseId: null
number: null
url: null
title: "fetch: pins: make an unconfirmed pin a conflict of its repository"
issueType: Bug
labels:
  - "scope: code"
  - "scope: test"
  - "type: bug"
  - "severity: major"
  - "impact: medium"
  - "effort: easy"
assignees:
  - kamilsk
milestone: null
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

# fetch: pins: make an unconfirmed pin a conflict of its repository

## Problem

When GitHub does not confirm the repository of a `repos` pin, for example a
deleted one, `fetch` prints no plan and stops with `cannot verify explicit
checkout`. One stale pin blocks every other repository of the run.
`docs/content/fetch.md` describes a pin failure as a conflict of that repository
only.

## PoC

```bash
# fetch.toml: a [[repos]] pin of a repository that GitHub answers 404 for,
# and a profile that discovers other repositories
maintainer fetch; echo "exit=$?"
# today:    "cannot verify explicit checkout", no plan
# expected: the plan of the other repositories and a conflict row for the pin
```

## Expected behavior

A pin that GitHub does not confirm is a conflict of its repository, with the pin
and the answer of GitHub. The rest of the plan is built and applied as usual,
and the exit code treats this conflict as any other conflict.

## Acceptance criteria

- [ ] A service test with a pin answered `404 Not Found` and one discovered
      repository plans the discovered repository and reports the pin as a
      conflict.
- [ ] The human and the JSON plans show the conflict with its reason.
- [ ] The output matches the pin failure that `docs/content/fetch.md` describes.

## Out of scope

Failures that are not about the pinned repository, such as a network error or a
rate limit: they keep their current handling.

Source: [20261125T152232Z](../reviews/20261125T152232Z.md), "A pin that GitHub does not confirm stops the whole fetch run".
