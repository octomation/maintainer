---
code: MAIN-112
id: I_kwDOE2M9Zc5tCLYi
databaseId: 1829287458
number: 153
url: https://github.com/octomation/maintainer/issues/153
title: "template: go new support"
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
createdAt: 2023-07-31T14:14:42Z
updatedAt: 2026-11-20T15:42:51Z
lastEditedAt: 2026-11-20T15:42:27Z
closedAt: null
---

# template: go new support

## Motivation

Starting a project from a template is manual: copy the files, rename the
module, fix the imports, and hope that a search and replace left no old path
behind. The [Go blog post](https://go.dev/blog/gonew) introduced `gonew`, a
separate tool from the [tools v0.11.1 release](https://github.com/golang/tools/releases/tag/v0.11.1)
that does this for Go modules, rather than a built-in `go new` command:

```bash
gonew golang.org/x/example/helloserver example.com/myserver
```

The idea was to combine it with [GitHub template repositories](https://docs.github.com/en/repositories/creating-and-managing-repositories/creating-a-template-repository)
and the octomation templates (go-module, go-tool, go-service), and later to
port the approach to other stacks for a better developer experience.

## PoC

```bash
$ maintainer template new octomation/go-tool@v1.2.3 github.com/kamilsk/foo ./foo
# ./foo: created from octomation/go-tool@v1.2.3 as github.com/kamilsk/foo
# next: git init, create the repository on GitHub, push
```

## Acceptance criteria

- [ ] The template source and its version are chosen explicitly, and the new
  project records them, so that [[MAIN-10]] knows the point to sync from.
- [ ] The module name and every import of the template module are rewritten:
  `go build ./...` passes in the new project, and no old path is left.
- [ ] A non-empty target directory is never overwritten implicitly.
- [ ] The output names the next steps.
- [ ] Only stacks with a working adapter are advertised; Go comes first.

## Out of scope

- Renaming the module of an existing project: [[MAIN-108]].
- Syncing a created project with its template later: [[MAIN-10]].
- Creating the GitHub repository itself.

Related: [[MAIN-108]], [[MAIN-10]].

<!-- 2023-08-09T12:36Z https://github.com/octomation/maintainer/issues/153#issuecomment-1671245480
related to #135
-->

<!-- 2023-08-26T19:26Z https://github.com/octomation/maintainer/issues/153#issuecomment-1694486616
- [ ] investigate how does it change a module name https://github.com/golang/tools/blob/master/cmd/gonew/main.go
-->

<!-- 2026-11-20T21:00Z https://github.com/octomation/maintainer/issues/153#issuecomment-7043759211
Maturity: idea. Nothing in code: the root command registers fetch, status, github, service and skill (internal/command/root.go), there is no template or new command, and git history has no attempt.
`gonew` and GitHub's "Use this template" (`gh repo create --template`) already cover the job.
Proposal for v0.5.0: not planned. Scaffolding source code from a stack template is a build helper, not work with repositories, issue trackers or GitHub, the same kind as `go vanity build` (MAIN-161) and `makefile build` (MAIN-162) that this release removes.
-->
