---
code: MAIN-4
id: MDU6SXNzdWU3Nzc0Mjg4MzM=
databaseId: 777428833
number: 4
url: https://github.com/octomation/maintainer/issues/4
title: "how to warmup godoc and vanity url"
issueType: null
labels: []
assignees:
  - kamilsk
milestone: "[[MAIN-M1]]"
parent: null
issueFields: {}
state: CLOSED
stateReason: NOT_PLANNED
duplicateOf: null
createdAt: 2021-01-02T08:55:25Z
updatedAt: 2026-11-21T05:33:28Z
lastEditedAt: 2026-11-21T05:18:06Z
closedAt: 2026-11-21T05:33:28Z
---

# how to warmup godoc and vanity url

## Motivation

After a Go module or a new subpackage is released, its vanity import path has to
resolve and its documentation has to appear on pkg.go.dev. `go vanity build`
([[MAIN-1]]) only writes the pages. That does not prove that `go get` reaches the
published path, that the Go proxy serves the version, or that the package index
lists it. The author wanted one step after a release that checks all three and
nudges the index.

The original idea called `POST https://godoc.org/-/refresh` with a `path`
parameter. godoc.org is gone, so that is not a live contract. pkg.go.dev
documents [adding a package](https://pkg.go.dev/about#adding-a-package): request
the page, or fetch the version through the Go proxy. Neither promises instant
indexing.

## PoC

The command name is provisional:

```bash
maintainer go vanity warmup go.octolab.org/toolkit/cli@v0.6.4
# vanity: reachable
# module version: available through the Go proxy
# documentation: requested; its appearance is verified separately
```

## Acceptance criteria

- [ ] The vanity page, the module version through the Go proxy and the
      documentation are separate stages, and the output names the stage that
      did not complete.
- [ ] The documentation is requested only through the interfaces pkg.go.dev
      documents; nothing calls the godoc.org endpoint.
- [ ] An unreachable path or a version the proxy does not serve is an error with
      a non-zero exit.
- [ ] A re-run after a new subpackage is published stays valid and checks the
      new path too.

## Out of scope

- Generating the vanity pages: [[MAIN-21]], [[MAIN-107]].
- A promise of when pkg.go.dev shows the page.

Related: [[MAIN-161]].

<!-- 2021-01-02T14:36Z https://github.com/octomation/maintainer/issues/4#issuecomment-753481151
https://go.dev/about#adding-a-package
-->

<!-- 2026-11-20T21:00Z https://github.com/octomation/maintainer/issues/4#issuecomment-7043717334
Maturity: idea. Nothing was built: `go vanity` only ever had `build` (`internal/command/golang/vanity.go`), and the whole `internal/command/golang` and `internal/model/golang` went in commit 1f95e2d (MAIN-161).
The need moved with the catalog: octolab/vanity checks the live site itself (`bin/verify.py --url`, and `bin/lib/core/check.bash` downloads modules through `GOPROXY=direct` with an empty cache), and octomation/vanity-action publishes each release from the module's own workflow.
v0.5.0 proposal: not planned. Checking the Go proxy and pkg.go.dev after a release is release tooling for one vanity host, not work with repositories, issue trackers or GitHub; if it is still wanted, it belongs to octolab/vanity or vanity-action.
-->

<!-- 2026-11-21T05:30Z https://github.com/octomation/maintainer/issues/4#issuecomment-7057978370
Closed as not planned: this is no longer the focus of maintainer. From v0.5.0 maintainer is the contributor's central tool: one picture of the repositories, issue trackers and GitHub across all projects, not build or release tooling.
Checking the Go proxy and pkg.go.dev after a release stays with octolab/vanity (`bin/verify.py --url`); the command it would extend went with #536.
-->
