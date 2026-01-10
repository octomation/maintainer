---
code:
id: MDU6SXNzdWU4MjM2ODc5MjU=
databaseId: 823687925
number: 11
url: https://github.com/octomation/maintainer/issues/11
title: "integrate markdown parser"
labels: []
milestone:
state: OPEN
stateReason:
createdAt: 2021-03-06T16:28:34Z
updatedAt: 2023-03-31T15:36:45Z
lastEditedAt:
closedAt:
---

# integrate markdown parser

Add structural processing of Markdown to extract project information and validate documentation. The maintainer needs the links to mirrors, quality reports and other integrations that are already present in the README, including those written as reference definitions for badges.

An example based on the current README:

```text
Input:      [![Mirror][mirror.icon]][mirror.page]
Definition: mirror.page → https://bitbucket.org/kamilsk/maintainer
Result:     the link to the repository mirror
```

Headings, links and code blocks must be distinguishable, `[name][ref]` links must resolve correctly, and missing mandatory sections must be reported. Text inside a code example must not become project metadata.

The historical candidate is [blackfriday](https://github.com/russross/blackfriday); choosing a library is not the user-facing outcome of this task. **At present** there is no Markdown parser among the dependencies; the existing HTML parser serves GitHub contributions. The motivation and the storage options are covered in [Repository Metadata](<../notes/Repository Metadata.md>); related scenarios are [#2](issue-2.md) and [#5](issue-5.md).
