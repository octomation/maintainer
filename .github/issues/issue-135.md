---
id: 135
database_id: 1672920804
node_id: I_kwDOE2M9Zc5jtr7k
status: open
title: "go: rename: command to rename current module and others"
labels: ["scope: code","type: feature","impact: high","effort: hard"]
url: https://github.com/octomation/maintainer/issues/135
created_at: 2023-04-18T11:32:08Z
updated_at: 2023-04-18T11:32:09Z
---

# go: rename: command to rename current module and others

Automate renaming the project's own Go module and the imports of its dependencies. This cuts the manual work after a project is created from a template such as [go-module](https://github.com/octomation/go-module), and when a dependency moves to a new major import path.

The provisional interface from the original idea:

```bash
maintainer go rename mod github.com/new/owner
maintainer go rename github.com/vendor/pkg github.com/vendor/pkg/v2
```

The first scenario changes the module path of the project and its internal imports; the second changes a chosen dependency path. Subpackages must keep their relative path, and similar strings in ordinary text must not be replaced blindly. A reviewable diff, clear boundaries for the affected modules, and a build check of the result are required.

**At present** there is no `go rename` command, although `tools/` already pins the helper tools `mod`, `gomvpkg` and `gorename`. Their presence provides no user-facing interface.

The original candidates: [mod](https://github.com/marwan-at-work/mod), [go-imports-rename](https://github.com/sirkon/go-imports-rename), [gopls](https://github.com/golang/tools/tree/master/gopls), [gorename](https://github.com/golang/tools/tree/master/cmd/gorename). Related: [#153](issue-153.md), the [dependency-update idea](<../notes/Maintainer draft idea.md>).
