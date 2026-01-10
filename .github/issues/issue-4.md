---
code:
id: MDU6SXNzdWU3Nzc0Mjg4MzM=
databaseId: 777428833
number: 4
url: https://github.com/octomation/maintainer/issues/4
title: "how to warmup godoc and vanity url"
labels: []
milestone:
state: OPEN
stateReason:
createdAt: 2021-01-02T08:55:25Z
updatedAt: 2023-01-06T13:48:58Z
lastEditedAt:
closedAt:
---

# how to warmup godoc and vanity url

After releasing a Go module or a new subpackage, verify that the vanity path resolves and trigger the appearance of documentation. Today `go vanity build` only produces pages: successful generation does not yet mean the published import path is reachable for users or listed in the package index.

Proposed scenario; the command name is provisional:

```bash
maintainer go vanity warmup go.octolab.org/toolkit/cli@v0.6.4
# vanity: reachable
# module version: available through the Go proxy
# documentation: requested; its appearance is verified separately
```

Checking the vanity page, the availability of the module version, and the readiness of documentation must be distinguishable, and the stage that did not complete must be visible. Re-running has to stay valid after a new subpackage is published.

The historical example used `POST https://godoc.org/-/refresh` with a `path` parameter. That cannot be treated as a live contract: [pkg.go.dev documents](https://pkg.go.dev/about#adding-a-package) requesting the page and loading the version through the Go proxy. The warm-up method should be chosen against those interfaces, without promising instant indexing.

<!-- 2021-01-02T14:36Z https://github.com/octomation/maintainer/issues/4#issuecomment-753481151
https://go.dev/about#adding-a-package
-->
