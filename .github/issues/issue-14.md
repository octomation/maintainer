---
id: 14
database_id: 841022494
node_id: MDU6SXNzdWU4NDEwMjI0OTQ=
status: closed
title: "add dry-run option for github labels"
labels: ["scope: code"]
url: https://github.com/octomation/maintainer/issues/14
created_at: 2021-03-25T15:04:20Z
updated_at: 2023-03-31T15:32:48Z
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

**Current state:** the issue is closed; the label commands were removed by [#80](issue-80.md). The example explains the original interface, not a current capability. The more general idea of describing a "before → after" transition keeps its meaning in [#19](issue-19.md).
