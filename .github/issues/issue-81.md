---
code:
id: I_kwDOE2M9Zc5PshxL
databaseId: 1337072715
number: 81
url: https://github.com/octomation/maintainer/issues/81
title: "makefile: build: allow include recursively"
labels:
  - "scope: code"
  - "scope: test"
milestone:
state: OPEN
stateReason:
createdAt: 2022-08-12T11:11:50Z
updatedAt: 2026-09-25T05:13:56Z
lastEditedAt: 2026-09-25T05:13:55Z
closedAt:
---

# makefile: build: allow include recursively

Let a Makefile fragment include another fragment explicitly, so the dependency does not have to be duplicated at the top level. The original example: `build.service.mk` uses `build.tool.mk`, but the relation was declared only in `Go Service.mk`.

A prototype of the input:

```make
# build.service.mk
include build.tool.mk
```

`maintainer makefile build 'Go Service.mk'` is expected to produce a self-contained result carrying the required rules in the right order. Path resolution, repeated includes, optional files and cycle diagnostics all have to be defined, so that recursion does not turn into a hang.

Closing this issue should mean verifying the original fragment structure and the missing cases, not redefining its goal as the mere presence of a recursive call.

<!-- 2023-08-09T12:38Z https://github.com/octomation/maintainer/issues/81#issuecomment-1671248449
I have to avoid recursive including and deep-nesting cases.
-->
