---
id: 135
database_id: 1672920804
node_id: I_kwDOE2M9Zc5jtr7k
status: open
title: "go: rename: command to rename current module and others"
labels: [scope: code, type: feature, impact: high, effort: hard]
url: https://github.com/octomation/maintainer/issues/135
created_at: 2023-04-18T11:32:08Z
updated_at: 2023-04-18T11:32:09Z
---

# go: rename: command to rename current module and others

**Motivation:** allow to use templates, such as [go-module](https://github.com/octomation/go-module), quicker.

**Details**

```bash
$ maintainer go rename mod github.com/new/owner

$ maintainer go rename github.com/vendor/pkg github.com/vendor/pkg/v2
```

**Research**

- https://github.com/marwan-at-work/mod
- https://github.com/sirkon/go-imports-rename
- https://github.com/golang/tools/tree/master/gopls
- https://github.com/golang/tools/tree/master/cmd/gorename
