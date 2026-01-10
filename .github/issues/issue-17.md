---
code:
id: MDU6SXNzdWU5NzkwODgzMjM=
databaseId: 979088323
number: 17
url: https://github.com/octomation/maintainer/issues/17
title: "extend gh tool"
labels: []
milestone:
state: CLOSED
stateReason: COMPLETED
createdAt: 2021-08-25T12:28:31Z
updatedAt: 2023-01-15T09:34:08Z
lastEditedAt:
closedAt: 2023-01-15T09:34:07Z
---

# extend gh tool

Research delivering maintainer's features as a GitHub CLI extension. A user who already works with `gh` should run the assistant in a familiar environment instead of managing a separate set of unrelated commands.

A provisional scenario:

```bash
gh maintainer contribution lookup 2021-02-01/3
```

The outcome of the research is a reproducible way to install and run the extension, plus decisions about which commands it exposes and where authorization comes from. How arguments and the stdout/stderr split survive being forwarded into maintainer needs checking.

**Current state:** the issue is closed, but no ready extension exists in the project tree. The later open task is [#97](issue-97.md), the dashboard experiment is [#53](issue-53.md).

Original material: [the extensions announcement](https://github.blog/2021-08-24-github-cli-2-0-includes-extensions/) and [scripting with gh](https://github.blog/2021-03-11-scripting-with-github-cli/).

<!-- 2023-01-15T09:34Z https://github.com/octomation/maintainer/issues/17#issuecomment-1383102532
duplicated by #97
-->
