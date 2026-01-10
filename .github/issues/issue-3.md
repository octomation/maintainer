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
updatedAt: 2026-09-25T05:11:38Z
lastEditedAt: 2026-09-25T05:11:38Z
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

Related: formatting of the result [[issue-7]], nested fragment dependencies [[issue-81]], original [octomation/makefiles#22](https://github.com/octomation/makefiles/issues/22).
