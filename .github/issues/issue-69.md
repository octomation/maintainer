---
code:
id: I_kwDOE2M9Zc5OQMi4
databaseId: 1312868536
number: 69
url: https://github.com/octomation/maintainer/issues/69
title: "boilerplate: fetch latest changes from go-tool before publishing v0.1.0"
labels:
  - "type: feature"
  - "scope: inventory"
  - "impact: medium"
  - "effort: easy"
milestone: "[[milestone-1]]"
state: OPEN
stateReason:
createdAt: 2022-07-21T08:20:55Z
updatedAt: 2023-04-06T11:16:17Z
lastEditedAt:
closedAt:
---

# boilerplate: fetch latest changes from go-tool before publishing v0.1.0

Carry the go-tool changes that were current at release time into maintainer before publishing v0.1.0. This is a one-off preparation of the release infrastructure: the shared template should provide a consistent build, consistent checks and consistent artifact delivery.

The result is a reviewable diff of the service files with the project's specifics preserved. For example: update the shared workflow configuration without losing the contribution tests or maintainer's publishing parameters.

**Reassessment:** the issue is open in the archive and is tied to preparing release [#30](issue-30.md). The original phrase `blocker is .../30` is ambiguous about the direction of the dependency; it must not turn into the circular requirement of "publish the release first, then prepare its template".

Before closing, it has to be decided which changes still belong to this historical release. The later synchronization is [#175](issue-175.md), and the general automated mechanism is [#10](issue-10.md).
