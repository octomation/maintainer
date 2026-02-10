---
code: MAIN-163
id: I_kwDOE2M9Zc8AAAABhOjypw
databaseId: 6524826279
number: 538
url: https://github.com/octomation/maintainer/issues/538
title: "cli: drop the proxy command, the unused mocks and the boilerplate feature"
issueType: Task
labels:
  - "scope: code"
  - "type: improvement"
  - "scope: deps"
assignees:
  - kamilsk
milestone: null
parent: null
issueFields: {}
state: OPEN
stateReason: null
duplicateOf: null
createdAt: 2026-11-21T05:16:46Z
updatedAt: 2026-11-21T05:17:17Z
lastEditedAt: null
closedAt: null
---

# cli: drop the proxy command, the unused mocks and the boilerplate feature

## Motivation

v0.5.0 slims maintainer down to its mission: one place for a contributor's work with
repositories, GitHub and other trackers. A survey of the code (`deadcode ./...` and the
importers of each package) found pieces that serve nothing:

- `internal/command/proxy` (2021, "hold proxy commands in separated package") is a bare
  `diff` wrapper: it passes the arguments through and swallows the exit status. It was
  never wired into the root command.
- `internal/command/github/contract.go` declares `Git` and `GitHub` interfaces only for
  `mocks_test.go`, and no test uses those mocks; no test uses the mocks of
  `internal/service/git` either. They are the only users of `github.com/golang/mock`
  (archived in 2023) and of the `mockgen` tool.
- The `boilerplate` feature of `internal/config/features.go` is a leftover of the project
  template: it enables nothing, but `maintainer version` prints it as enabled.

The rest of the unwired code stays on purpose: `hub`, `github/management`,
`service/git/provider`, the `bitbucket`, `gitlab` and `git` models and `service/git` are
the domain of maintainer for the issue and repository work to come, and `internal/pkg/*`
are candidates to move to go.octolab.org.

## PoC

```bash
go run golang.org/x/tools/cmd/deadcode@latest ./... | grep -E 'proxy|contract'
grep -rn 'golang/mock' go.mod tools/go.mod
go run . version | grep features
```

## Acceptance criteria

- [x] `internal/command/proxy` is gone.
- [x] `internal/command/github/contract.go` and both `mocks_test.go` are gone, with the
      `mockgen` lines of `internal/service/git/contract.go`; `github.com/golang/mock` is in
      neither `go.mod` nor `tools/go.mod`, and `tools/README.md` no longer lists `mockgen`.
- [x] The `boilerplate` feature is gone, `maintainer version` prints `features : -`; the
      feature list of `internal/config/features.go` stays for the features to come.
- [x] `make lint`, `go test ./...` and `make go-generate` pass and leave the tree clean.
- [ ] The changelog of v0.5.0 names the removal of the dependency.

## Out of scope

- `hub`, `github/management`, `service/git/provider`, the `bitbucket`, `gitlab` and `git`
  models, `service/git` and `internal/pkg/*`: kept, see Motivation.
- Dead functions inside live packages (`pkg/time`, `service/fetch`, `skill`, …): a
  separate pass, one by one, as some are used by tests.

<!-- 2026-11-20T21:30Z https://github.com/octomation/maintainer/issues/538#issuecomment-7057807892
Prepared in the working tree for v0.5.0; local only, not on GitHub yet.

- Removed: `internal/command/proxy/root.go`, `internal/command/github/contract.go`, `internal/command/github/mocks_test.go`, `internal/service/git/mocks_test.go`, the two `go:generate` lines of `internal/service/git/contract.go`.
- `go.mod` loses `github.com/golang/mock v1.6.0`; `tools/go.mod` loses the `mockgen` tool and `golang/mock v1.7.0-rc.1`; `tools/README.md` keeps `goimports` for `make format` only.
- `internal/config/features.go`: the `boilerplate` entry is gone, the empty `Features` list stays; `go run . version` prints `features : -`.
- `.github/self-control.md`: the proxy and hub item is resolved.
- Checks: `go build`, `go vet`, `go test ./...`, `make lint` (0 issues), `make go-generate` (tree unchanged).
-->
