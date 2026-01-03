---
id: 23
database_id: 1216401742
node_id: I_kwDOE2M9Zc5IgNFO
status: closed
title: "sync with latest template"
labels: []
url: https://github.com/octomation/maintainer/issues/23
created_at: 2022-04-26T19:44:40Z
updated_at: 2022-05-06T14:07:27Z
---

# sync with latest template

Synchronize maintainer itself with its source template, go-tool, as a one-off. The goal is to pick up shared improvements to the build, the checks and the service configuration while keeping the specifics of this CLI.

The result has to be presented as a reviewable diff: what came from the template, which local settings were kept, and which divergences need a separate decision. After the transfer, building the binary and the existing commands must still work.

**Current state:** the historical task is closed; the [README](../../README.md) still names go-tool as the project template. That is not evidence that the repository matches today's version of the template. Later one-off updates are considered in [#69](issue-69.md) and [#175](issue-175.md), automation of the process in [#10](issue-10.md).
