---
code:
id: MDU6SXNzdWU3NzY2MDI5MTU=
databaseId: 776602915
number: 1
url: https://github.com/octomation/maintainer/issues/1
title: "import go code from vanity project"
labels:
  - "help wanted"
milestone:
state: CLOSED
stateReason: COMPLETED
createdAt: 2020-12-30T19:25:12Z
updatedAt: 2021-01-01T18:25:16Z
lastEditedAt:
closedAt: 2021-01-01T18:25:16Z
---

# import go code from vanity project

Move Go vanity page generation out of the separate `octomation/vanity` project into maintainer. This consolidates the maintenance tooling: a project's own import path can be built with the same CLI as the rest of its artifacts.

Expected scenario: the maintainer describes the modules and their repositories, then gets HTML pages carrying import metadata and links to sources.

```bash
maintainer go vanity build --file modules.yml --host go.octolab.org dist
```

The result is a tree of pages under `dist`, ready to be published separately on the given domain. The command does not configure DNS or hosting.

**Current state:** the command and the generator are present in [the code](../../internal/command/golang/vanity.go). The historical task is closed; follow-up work on complex module layouts is [#21](issue-21.md), automatic discovery is [#134](issue-134.md). Original link: [octomation/vanity#6](https://github.com/octomation/vanity/issues/6).
