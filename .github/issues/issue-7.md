---
code:
id: MDU6SXNzdWU4MTExMTAwNzA=
databaseId: 811110070
number: 7
url: https://github.com/octomation/maintainer/issues/7
title: "remove double new lines"
labels:
  - "help wanted"
milestone:
state: CLOSED
stateReason: COMPLETED
createdAt: 2021-02-18T13:41:46Z
updatedAt: 2021-03-05T19:07:21Z
lastEditedAt:
closedAt: 2021-03-05T19:07:21Z
---

# remove double new lines

Remove the redundant blank lines that appear when `include` directives are expanded during a Makefile build. The result must stay readable and must not produce noisy changes caused only by the boundaries between merged fragments.

Input example:

```make
include src/common/env.mk

export PATH := $(GOBIN):$(PATH)
```

After the contents of `env.mk`, exactly one blank separator is expected before `export`, even when the included file already ends with a blank line. Recipe tabs and non-empty lines are preserved.

**Clarification of the original report:** its `expected` and `obtained` blocks contradicted the title — two blank lines were labelled as expected. In the current [assembler](../../internal/command/makefile/pkg.go) a run of newlines is capped at two characters, that is one blank line. The closed task describes exactly this removal of duplicates.
