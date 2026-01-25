---
code: MAIN-13
id: MDU6SXNzdWU4MzM2MTMxODA=
databaseId: 833613180
number: 13
url: https://github.com/octomation/maintainer/issues/13
title: "label presets"
issueType: null
labels:
  - "scope: code"
assignees:
  - kamilsk
milestone: null
parent: null
issueFields: {}
state: CLOSED
stateReason: COMPLETED
duplicateOf: null
createdAt: 2021-03-17T10:16:07Z
updatedAt: 2026-09-25T05:10:09Z
lastEditedAt: 2026-09-25T05:10:09Z
closedAt: 2021-03-25T11:31:31Z
---

# label presets

Provide label presets so that repositories can be brought to a common system of categories with a manual review step. A preset defines the wanted names, colors and descriptions, and the maintainer sees the result before it is applied.

The historical prototype:

```bash
maintainer github labels dump --preset=octolab > .git/labels.yml
# review and, if needed, edit .git/labels.yml
maintainer github labels update < .git/labels.yml
```

A clear mapping of the existing labels onto the preset categories is expected — the original sketch described it as a classification tree whose most appropriate node supplies the transform rule. Ambiguous cases must be visible in the proposal, and reprocessing an already conforming set must not create new changes.

Related: umbrella [[MAIN-8]], preview [[MAIN-14]], wider repository set [[MAIN-15]], label scope removal [[MAIN-80]].
