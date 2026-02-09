---
code: MAIN-107
id: I_kwDOE2M9Zc5jonsA
databaseId: 1671592704
number: 134
url: https://github.com/octomation/maintainer/issues/134
title: "go: vanity: build from git submodules and go.work instead of modules.yml"
issueType: null
labels:
  - "scope: code"
  - "type: feature"
  - "impact: high"
  - "effort: medium"
assignees:
  - kamilsk
milestone: "[[MAIN-M1]]"
parent: null
issueFields: {}
state: OPEN
stateReason: null
duplicateOf: null
createdAt: 2023-04-17T16:51:00Z
updatedAt: 2026-11-20T15:42:19Z
lastEditedAt: 2026-11-20T15:42:19Z
closedAt: null
---

# go: vanity: build from git submodules and go.work instead of modules.yml

## Motivation

`go vanity build` read the list of modules from a hand-written `modules.yml`.
The same list already lives in `go.work` and in the Git submodules of the
repository, so every added or moved module had to be edited in several places.
The goal: discover the modules from those sources, so the next build reflects a
change with no manual step.

The precedence of the sources and compatibility with an existing `modules.yml`
have to be decided before implementing.

## PoC

```text
go.work: use ./toolkit, ./services/api
Git submodules: the directories holding module sources
→ the discovered module paths and their sources
→ vanity pages for those modules and packages
```

```bash
maintainer go vanity build --discover --dry-run
# go.octolab.org/toolkit       ./toolkit         go.work
# go.octolab.org/services/api  ./services/api    go.work, submodule
```

## Acceptance criteria

- [ ] The discovered set is shown before anything is published.
- [ ] The output tells a Go module from a Git submodule, and names the source
      of each entry.
- [ ] A missing directory and two modules with one import path are errors that
      name them.
- [ ] An explicit exclusion and a directory that does not match its import path
      stay configurable.
- [ ] A module added to `go.work` appears in the next build with no other edit.

## Out of scope

- Several modules and mismatched paths in the description itself: [[MAIN-21]].
- Creating a project from a template: [[MAIN-112]].

Related: [[MAIN-161]].

<!-- 2026-11-20T21:00Z https://github.com/octomation/maintainer/issues/134#issuecomment-7043712515
Maturity: idea. `go vanity build` only decoded the YAML named by `--file` (`internal/command/golang/vanity.go`); there was never any `go.work` or submodule discovery. The command went in commit 1f95e2d (MAIN-161).
There is no central list left to discover: octomation/vanity-action runs in each module's release workflow and lists that module's packages at the released commit.
v0.5.0 proposal: not planned. Module discovery for vanity pages is a Go build helper off the mission, and the push model of vanity-action removes the duplicate list it was meant to replace.
-->
