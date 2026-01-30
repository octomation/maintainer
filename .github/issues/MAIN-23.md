---
code: MAIN-23
id: I_kwDOE2M9Zc5IgNFO
databaseId: 1216401742
number: 23
url: https://github.com/octomation/maintainer/issues/23
title: "sync with latest template"
issueType: null
labels: []
assignees:
  - kamilsk
milestone: "[[MAIN-M1]]"
parent: null
issueFields: {}
state: CLOSED
stateReason: COMPLETED
duplicateOf: null
createdAt: 2022-04-26T19:44:40Z
updatedAt: 2026-10-20T07:32:02Z
lastEditedAt: 2026-09-25T05:11:18Z
closedAt: 2022-05-06T14:07:27Z
---

# sync with latest template

Synchronize maintainer itself with its source template, go-tool, as a one-off. The goal is to pick up shared improvements to the build, the checks and the service configuration while keeping the specifics of this CLI.

The result has to be presented as a reviewable diff: what came from the template, which local settings were kept, and which divergences need a separate decision. After the transfer, building the binary and the existing commands must still work.

Related: [[MAIN-69]], [[MAIN-116]], [[MAIN-10]].
