---
code: MAIN-108
id: I_kwDOE2M9Zc5jtr7k
databaseId: 1672920804
number: 135
url: https://github.com/octomation/maintainer/issues/135
title: "go: rename: command to rename current module and others"
issueType: null
labels:
  - "scope: code"
  - "type: feature"
  - "impact: high"
  - "effort: hard"
assignees:
  - kamilsk
milestone: "[[MAIN-M1]]"
parent: null
issueFields: {}
state: OPEN
stateReason: null
duplicateOf: null
createdAt: 2023-04-18T11:32:08Z
updatedAt: 2026-10-20T07:30:53Z
lastEditedAt: 2026-10-19T07:22:26Z
closedAt: null
---

# go: rename: command to rename current module and others

Automate renaming the project's own Go module and the imports of its dependencies. This cuts the manual work after a project is created from a template such as [go-module](https://github.com/octomation/go-module), and when a dependency moves to a new major import path.

The provisional interface from the original idea:

```bash
maintainer go rename mod github.com/new/owner
maintainer go rename github.com/vendor/pkg github.com/vendor/pkg/v2
```

The first scenario changes the module path of the project and its internal imports; the second changes a chosen dependency path. Subpackages must keep their relative path, and similar strings in ordinary text must not be replaced blindly. A reviewable diff, clear boundaries for the affected modules, and a build check of the result are required.

The original candidates: [mod](https://github.com/marwan-at-work/mod), [go-imports-rename](https://github.com/sirkon/go-imports-rename), [gopls](https://github.com/golang/tools/tree/master/gopls), [gorename](https://github.com/golang/tools/tree/master/cmd/gorename). Related: [[MAIN-112]].
