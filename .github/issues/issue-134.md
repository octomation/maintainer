---
code:
id: I_kwDOE2M9Zc5jonsA
databaseId: 1671592704
number: 134
url: https://github.com/octomation/maintainer/issues/134
title: "go: vanity: build from git submodules and go.work instead of modules.yml"
labels:
  - "scope: code"
  - "type: feature"
  - "impact: high"
  - "effort: medium"
milestone:
state: OPEN
stateReason:
createdAt: 2023-04-17T16:51:00Z
updatedAt: 2023-04-17T16:51:00Z
lastEditedAt:
closedAt:
---

# go: vanity: build from git submodules and go.work instead of modules.yml

Discover the Go modules for a vanity site from Git submodules and `go.work`, so that a duplicate list does not have to be maintained by hand in `modules.yml` — better automation, less manual work. Adding or moving a module should be reflected in the next generation without editing several sources.

The proposed scenario:

```text
go.work: use ./toolkit, ./services/api
Git submodules: the directories holding module sources
→ the discovered module paths and their sources
→ vanity pages for those modules and packages
```

The discovered set has to be shown before publication, a Go module has to be distinguishable from a Git submodule, and missing directories and conflicting import paths have to be diagnosed. Explicit exclusions, and a directory that does not match its import path, must stay manageable.

**At present** `maintainer go vanity build` reads only the `--file` file, `modules.yml` by default; there is no discovery. Before implementing, the precedence of sources and compatibility with the existing configuration have to be decided. Complex path mappings and multiple modules are handled in [#21](issue-21.md); creating a project from a template is [#153](issue-153.md).
