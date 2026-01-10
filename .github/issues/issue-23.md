---
code:
id: I_kwDOE2M9Zc5IgNFO
databaseId: 1216401742
number: 23
url: https://github.com/octomation/maintainer/issues/23
title: "sync with latest template"
labels: []
milestone:
state: CLOSED
stateReason: COMPLETED
createdAt: 2022-04-26T19:44:40Z
updatedAt: 2022-05-06T14:07:27Z
lastEditedAt:
closedAt: 2022-05-06T14:07:27Z
---

# sync with latest template

Synchronize maintainer itself with its source template, go-tool, as a one-off. The goal is to pick up shared improvements to the build, the checks and the service configuration while keeping the specifics of this CLI.

The result has to be presented as a reviewable diff: what came from the template, which local settings were kept, and which divergences need a separate decision. After the transfer, building the binary and the existing commands must still work.

**Current state:** the historical task is closed; the [README](../../README.md) still names go-tool as the project template. That is not evidence that the repository matches today's version of the template. Later one-off updates are considered in [#69](issue-69.md) and [#175](issue-175.md), automation of the process in [#10](issue-10.md).
