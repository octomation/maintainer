---
code:
id: MDU6SXNzdWU3NzcyOTU5NzE=
databaseId: 777295971
number: 3
url: https://github.com/octomation/maintainer/issues/3
title: "import go code from makefiles project"
labels:
  - "help wanted"
milestone:
state: CLOSED
stateReason: COMPLETED
createdAt: 2021-01-01T14:38:09Z
updatedAt: 2021-01-01T16:02:44Z
lastEditedAt:
closedAt: 2021-01-01T16:02:44Z
---

# import go code from makefiles project

Move the Makefile assembler out of `octomation/makefiles` into maintainer. Shared build fragments should live in one place, while consumers receive a self-contained Makefile without copying include files by hand.

An example of the working interface:

```bash
maintainer makefile build 'Go Tool.mk' 'Go Service.mk'
# dist/Go Tool/Makefile
# dist/Go Service/Makefile
```

The assembler expands includes and writes a separate result for every input file. The list of files can also be passed as lines on stdin.

**Current state:** the move is reflected in the [makefile commands](../../internal/command/makefile/). The task is closed. Formatting of the result is refined by [#7](issue-7.md), nested fragment dependencies by [#81](issue-81.md). Original link: [octomation/makefiles#22](https://github.com/octomation/makefiles/issues/22).
