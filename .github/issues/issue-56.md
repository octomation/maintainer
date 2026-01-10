---
code:
id: I_kwDOE2M9Zc5Lxl4c
databaseId: 1271291420
number: 56
url: https://github.com/octomation/maintainer/issues/56
title: "github: contribution: support --short for suggest"
labels:
  - "scope: docs"
  - "scope: code"
milestone: "[[milestone-1]]"
state: CLOSED
stateReason: COMPLETED
createdAt: 2022-06-14T19:54:39Z
updatedAt: 2022-06-15T10:51:18Z
lastEditedAt: 2022-06-15T10:51:18Z
closedAt: 2022-06-15T10:38:28Z
---

# github: contribution: support --short for suggest

Make the result of suggest convenient for shell substitution and Git wrappers. A script needs a single date, while a person running it interactively benefits from a table explaining the choice.

The working example:

```bash
timestamp=$(maintainer github contribution suggest --short 2021/+10)
printf '%s\n' "$timestamp"
```

`--short` drops the table. The date goes to stdout and the explanation `Suggestion is …, actual → target` to stderr, so it does not end up in the variable. `--delta` returns a relative value instead of an absolute timestamp.

**Current state:** the closed task is implemented in [suggest](../../internal/command/github/contribution/suggest.go). The original example with an elaborate Git alias is replaced by direct shell substitution: `git at` and `git contribute` are external wrappers. Choosing the date from local history is related to [#58](issue-58.md).

<!-- 2022-06-15T10:38Z https://github.com/octomation/maintainer/issues/56#issuecomment-1156306472
https://github.com/octomation/maintainer/commit/8e9be2e7946be82a7f33ff0eb89f33ac37b3d6a2
-->
