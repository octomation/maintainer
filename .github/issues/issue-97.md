---
code:
id: I_kwDOE2M9Zc5bavOt
databaseId: 1533735853
number: 97
url: https://github.com/octomation/maintainer/issues/97
title: "github: cli: support extension mechanism"
labels:
  - "scope: code"
  - "type: feature"
  - "impact: high"
  - "effort: hard"
milestone:
state: OPEN
stateReason:
createdAt: 2023-01-15T09:33:02Z
updatedAt: 2026-09-25T05:14:24Z
lastEditedAt: 2026-09-25T05:14:24Z
closedAt:
---

# github: cli: support extension mechanism

Produce a verifiable way of distributing maintainer as a GitHub CLI extension, as an experiment in distribution and in the idiomatic way of interacting with GitHub. A `gh` user should install and run the features the familiar way, instead of wiring several binaries and scripts together by hand.

A provisional interface:

```bash
gh maintainer contribution lookup 2021-02-01/3
```

The name and packaging of the extension, argument forwarding, the source of authorization, and platform compatibility all have to be decided. Readiness means reproducible installation, update and execution of at least one contribution scenario, with its output and errors preserved.

This is the continuation of the research in [[issue-17]]; the dashboard [[issue-53]] may use the same delivery mechanism but is not required for a first result.

Original material: [extension tools](https://github.blog/2023-01-13-new-github-cli-extension-tools/), [the announcement of the mechanism](https://github.blog/2021-08-24-github-cli-2-0-includes-extensions/), [automation with gh](https://github.blog/2021-03-11-scripting-with-github-cli/).

<!-- 2023-01-15T09:35Z https://github.com/octomation/maintainer/issues/97#issuecomment-1383102775
After resolving the issue, I have to check https://www.infoq.com/news/2023/01/docker-desktop-416/.
-->
