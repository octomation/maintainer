---
code:
id: MDU6SXNzdWU4NDEwMjI0OTQ=
databaseId: 841022494
number: 14
url: https://github.com/octomation/maintainer/issues/14
title: "add dry-run option for github labels"
labels:
  - "scope: code"
milestone:
state: CLOSED
stateReason: NOT_PLANNED
createdAt: 2021-03-25T15:04:20Z
updatedAt: 2026-09-25T05:10:26Z
lastEditedAt: 2026-09-25T05:10:26Z
closedAt: 2023-03-31T15:32:48Z
---

# add dry-run option for github labels

Make it possible to review GitHub label changes without executing them. A bulk rename or deletion is hard to verify from the resulting set alone: the list of concrete actions is needed before any write operation is issued.

The historical prototype:

```bash
maintainer github labels update --dry-run < .git/labels.yml
# rename: bug → type: bug
# create: effort: easy
# delete: obsolete
```

The plan is expected to be exactly the one a normal run would apply, naming the affected labels and changing nothing on GitHub. "Show, don't do" — and the absence of changes must be reported explicitly.

Related: the general "before → after" transition [[issue-19]], label scope removal [[issue-80]].
