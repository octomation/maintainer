---
code:
id: MDU6SXNzdWU4MjM2ODY1MjM=
databaseId: 823686523
number: 9
url: https://github.com/octomation/maintainer/issues/9
title: "research: automate update golangci-lint config"
labels: []
milestone:
state: CLOSED
stateReason: NOT_PLANNED
createdAt: 2021-03-06T16:21:56Z
updatedAt: 2023-03-25T20:34:07Z
lastEditedAt:
closedAt: 2023-03-25T20:34:01Z
---

# research: automate update golangci-lint config

Research how to distribute a shared golangci-lint configuration across the Go templates. Identical rules in several repositories drift apart, and updating each file by hand increases the maintenance load.

The original sample: the configurations of [go-module](https://github.com/octomation/go-module/blob/master/.golangci.yml), [go-service](https://github.com/octomation/go-service/blob/master/.golangci.yml) and [go-tool](https://github.com/octomation/go-tool/blob/master/.golangci.yml).

The expected outcome of the research is a way to obtain the diff of the shared part, keep deliberate per-project exceptions, and propose an update. For example: "a rule was added to the template; it is absent in two projects; in the third it is overridden locally".

**Current state:** the issue is closed; there is no generic synchronizer in the CLI. The project's own `.golangci.yml` exists, but it is not evidence of automation. The general template synchronization scenario continues in [#10](issue-10.md).

<!-- 2023-03-25T20:34Z https://github.com/octomation/maintainer/issues/9#issuecomment-1483916166
won't do
-->
