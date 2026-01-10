---
code:
id: I_kwDOE2M9Zc5JgD4-
databaseId: 1233141310
number: 34
url: https://github.com/octomation/maintainer/issues/34
title: "disable hub subcommands"
labels:
  - "scope: code"
milestone:
state: CLOSED
stateReason: COMPLETED
createdAt: 2022-05-11T20:20:22Z
updatedAt: 2022-05-12T20:14:11Z
lastEditedAt:
closedAt: 2022-05-12T20:14:11Z
---

# disable hub subcommands

Remove the unfinished hub subcommands from maintainer's public interface. The user-facing help should list features that can actually be used; otherwise an experimental wrapper looks like a supported way of working with GitHub.

Expected observable behaviour: `maintainer --help` does not list `hub`, while the existing `github`, `go` and `makefile` groups remain available.

**Current state:** the issue is closed; the [root command](../../internal/command/root.go) does not register hub, although the experimental code under `internal/command/hub` and `proxy` is still in the tree. The public interface is defined by the absence of registration, not by deleting every source file.

The original point of discussion is the [historical root.go](https://github.com/octomation/maintainer/blob/5a6acac4d82aa3abe1985ce482c4e21417cb12bf/internal/command/root.go#L54). Experiments with `gh` extensions continue separately in [#97](issue-97.md).
