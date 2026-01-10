---
code:
id: MDU6SXNzdWU3ODM1OTc4ODc=
databaseId: 783597887
number: 6
url: https://github.com/octomation/maintainer/issues/6
title: "issue with modules without go files on the root"
labels: []
milestone:
state: CLOSED
stateReason: COMPLETED
createdAt: 2021-01-11T18:25:15Z
updatedAt: 2023-08-09T12:43:11Z
lastEditedAt:
closedAt: 2023-08-09T12:43:11Z
---

# issue with modules without go files on the root

Fix vanity page generation for Go modules that have no `.go` files at their root, with packages living deeper. A missing root package must not prevent the import path of an existing subpackage from resolving.

A minimal example of such a layout:

```text
module: example.org/tools
packages:
  example.org/tools/cli/flag
```

Pages are expected not only for `tools/cli/flag`, but also for the intermediate paths `tools` and `tools/cli`, carrying the metadata of the right repository. Repeated parents must not produce divergent results.

**Current state:** the issue is closed; [the intermediate-path generator and its tests](../../internal/model/golang/vanity/issue6_test.go) are named after this task. That does not cover the multiple-module and mismatched-path support from [#21](issue-21.md). Original report: [octomation/vanity#7](https://github.com/octomation/vanity/issues/7).

<!-- 2023-08-09T12:43Z https://github.com/octomation/maintainer/issues/6#issuecomment-1671256156
fixed
-->
