---
id: 13
database_id: 833613180
node_id: MDU6SXNzdWU4MzM2MTMxODA=
status: closed
title: "label presets"
labels: [scope: code]
url: https://github.com/octomation/maintainer/issues/13
created_at: 2021-03-17T10:16:07Z
updated_at: 2023-03-31T15:34:20Z
---

# label presets

Top level:

```bash
$ maintainer github labels dump --preset=octolab > .git/labels.yml
$ vim .git/labels.yml # to manual review and fixes
$ maintainer github labels update < .git/labels.yml
```

Low level:
- define classify tree
- find the most appropriate node
- find transform rule for preset
- replace it by the rule
- profit
