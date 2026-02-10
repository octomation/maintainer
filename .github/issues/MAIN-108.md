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
state: CLOSED
stateReason: NOT_PLANNED
duplicateOf: null
createdAt: 2023-04-18T11:32:08Z
updatedAt: 2026-11-21T05:33:39Z
lastEditedAt: 2026-11-20T15:42:21Z
closedAt: 2026-11-21T05:33:39Z
---

# go: rename: command to rename current module and others

## Motivation

A project created from a template such as
[go-module](https://github.com/octomation/go-module) starts with the template's
module path, and every import of its own packages has to be rewritten by hand. A
dependency that moves to a new major version needs the same rewrite of its
import path. The goal: one command for both.

The original candidates to build on or wrap:
[mod](https://github.com/marwan-at-work/mod),
[go-imports-rename](https://github.com/sirkon/go-imports-rename),
[gopls](https://github.com/golang/tools/tree/master/gopls),
[gorename](https://github.com/golang/tools/tree/master/cmd/gorename).

## PoC

The interface from the original idea, provisional:

```bash
maintainer go rename mod github.com/new/owner
maintainer go rename github.com/vendor/pkg github.com/vendor/pkg/v2
```

The first changes the module path of the project and its internal imports; the
second changes the path of a chosen dependency.

## Acceptance criteria

- [ ] After `go rename mod`, `go.mod` and every internal import carry the new
      path, and subpackages keep their relative path.
- [ ] After a dependency rename, only imports of that path and its subpackages
      change, and `go.mod` requires the new path.
- [ ] Similar strings in comments, strings and documentation are not replaced
      blindly.
- [ ] The change is shown as a reviewable diff, names the affected modules, and
      `go build ./...` passes on the result.

## Out of scope

- Creating a project from a template: [[MAIN-112]].

<!-- 2026-11-20T21:00Z https://github.com/octomation/maintainer/issues/135#issuecomment-7043713379
Maturity: idea. There is no `go rename` command and no `go` group any more (it went in commit 1f95e2d). `tools/go.mod` pins `mod`, `gomvpkg` and `gorename` through the tool directive, as developer tools of this repository only.
v0.5.0 proposal: not planned. Rewriting Go module and import paths is a source codemod, not work with repositories, issue trackers or GitHub; `mod` and gopls already do it.
-->

<!-- 2026-11-21T05:30Z https://github.com/octomation/maintainer/issues/135#issuecomment-7057975976
Closed as not planned: this is no longer the focus of maintainer. From v0.5.0 maintainer is the contributor's central tool: one picture of the repositories, issue trackers and GitHub across all projects, not build or release tooling.
`mod` and gopls rewrite Go sources. Renaming a module after a template is part of starting a new project, which #153 keeps open.
-->
