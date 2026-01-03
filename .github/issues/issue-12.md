---
id: 12
database_id: 826474736
node_id: MDU6SXNzdWU4MjY0NzQ3MzY=
status: closed
title: "fetch current state of labels of all repositories"
labels: ["help wanted"]
url: https://github.com/octomation/maintainer/issues/12
created_at: 2021-03-09T18:58:04Z
updated_at: 2023-03-31T15:34:19Z
---

# fetch current state of labels of all repositories

Obtain an overview of the current labels across the selected repositories, so that the real divergences are visible before standardization. A bare list of labels without their repository is not enough: identically named categories may carry different colors and descriptions.

A prototype of the ordered dump:

```text
Repository             Label          Color    Description
octomation/maintainer  type: bug      d73a4a   Something isn't working
kamilsk/retry          bug            d73a4a   Something isn't working
```

A stable order of repositories and labels is expected, preserving name, color and description. An empty set must be distinguishable from a failure to read the data. The dump feeds comparison and preset preparation; it changes nothing on GitHub by itself.

**Current state:** a historical, closed part of [#8](issue-8.md). After [#80](issue-80.md) there are no label commands in the CLI. The idea of a structured dump for external processing continues separately in [maintainer as raw](<../notes/Issues/maintainer as raw.md>).
