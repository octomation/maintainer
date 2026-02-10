---
code: MAIN-21
id: I_kwDOE2M9Zc49p2Bw
databaseId: 1034379376
number: 21
url: https://github.com/octomation/maintainer/issues/21
title: "improve vanity support"
issueType: null
labels:
  - "scope: code"
assignees:
  - kamilsk
milestone: "[[MAIN-M1]]"
parent: null
issueFields: {}
state: CLOSED
stateReason: NOT_PLANNED
duplicateOf: null
createdAt: 2021-10-24T10:30:31Z
updatedAt: 2026-11-21T05:33:31Z
lastEditedAt: 2026-11-21T05:18:04Z
closedAt: 2026-11-21T05:33:31Z
---

# improve vanity support

## Motivation

`go vanity build` ([[MAIN-1]]) wrote the vanity pages from a hand-written
`modules.yml`. That works for one module per repository whose import path
follows its directories. It is not enough for a complex layout, such as
qase-go: several Go modules in one repository, and directories published under
other paths.

The cases to support, from the original qase-go work
([dump](https://github.com/qase-tms/qase-go/commit/ffaccae028365f842cabc0bd63cb071171d2b47e),
[submodules](https://github.com/qase-tms/qase-go/commit/e12c8ca0c17082970e1af13bd11efb56d0408dfe)):

- `dump` the module description, so the data is visible before the pages are
  written;
- several modules inside one repository;
- a directory mapped to another public path:
  `reporters/<reporter>` → `go.qase.io/reporter/<reporter>`;
- source links through `blob` for a file and `tree` for a directory, instead of
  `tree` for both.

## PoC

```bash
maintainer go vanity dump --file modules.yml --host go.qase.io
# go.qase.io/reporter/ginkgo  github.com/qase-tms/qase-go  reporters/ginkgo
maintainer go vanity build --file modules.yml --host go.qase.io site/
curl -s https://go.qase.io/reporter/ginkgo | grep go-import
```

## Acceptance criteria

- [ ] `dump` prints every module with its import path, repository and source
      directory, and writes nothing.
- [ ] Every module of a repository gets its own pages, and its `go-import`
      names its own prefix.
- [ ] A directory published under another path resolves: `go get` of
      `go.qase.io/reporter/<reporter>` fetches `reporters/<reporter>`.
- [ ] A `go-source` link leads to the intended file through `blob` and to the
      intended directory through `tree`.

## Out of scope

- Subpackages without a root package: [[MAIN-6]].
- Discovering the modules automatically: [[MAIN-107]].
- Warming up the published paths: [[MAIN-4]].

Related: [[MAIN-161]].

<!-- 2026-11-20T21:00Z https://github.com/octomation/maintainer/issues/21#issuecomment-7043715370
Maturity: partly built, then removed. `--file` and `--host` came in commit 564d5bd (2021, issue #21); several modules with `import[].source` url/dir/file templates were readable (`internal/model/golang/entities.go`, `internal/model/golang/vanity/publisher.go`), but there was no `dump`, only the first import was used, and `fill` did not stop for a path outside the prefix.
All of it went in commit 1f95e2d (MAIN-161). octomation/vanity-action now runs in each module's own release and writes one `modules/<module path>.json` entry per module, so several modules in one repository publish one by one.
v0.5.0 proposal: not planned. Rendering vanity pages is a Go build helper off the mission; go.octolab.org renders them itself.
-->

<!-- 2026-11-21T05:30Z https://github.com/octomation/maintainer/issues/21#issuecomment-7057977935
Closed as not planned: this is no longer the focus of maintainer. From v0.5.0 maintainer is the contributor's central tool: one picture of the repositories, issue trackers and GitHub across all projects, not build or release tooling.
Modules now publish their own entries with octomation/vanity-action, and `go vanity build` went with #536.
-->
