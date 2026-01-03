---
id: 13
database_id: 833613180
node_id: MDU6SXNzdWU4MzM2MTMxODA=
status: closed
title: "label presets"
labels: ["scope: code"]
url: https://github.com/octomation/maintainer/issues/13
created_at: 2021-03-17T10:16:07Z
updated_at: 2023-03-31T15:34:20Z
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

**Current state:** the task is closed and the interface above is not available in the current version: the label scope was removed by [#80](issue-80.md). Presets relate to the umbrella task [#8](issue-8.md), the preview [#14](issue-14.md) and the wider repository set [#15](issue-15.md).
