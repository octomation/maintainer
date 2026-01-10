---
code:
id: I_kwDOE2M9Zc5LUx7T
databaseId: 1263738579
number: 51
url: https://github.com/octomation/maintainer/issues/51
title: "di: define service provider to inject it into command for lazy service initialization"
labels:
  - "scope: code"
  - "scope: test"
milestone:
state: OPEN
stateReason:
createdAt: 2022-06-07T19:01:10Z
updatedAt: 2026-09-25T05:12:35Z
lastEditedAt: 2026-09-25T05:12:35Z
closedAt:
---

# di: define service provider to inject it into command for lazy service initialization

Construct services only for the operation being executed, and only after its configuration is resolved. That makes it possible to test commands with substitutable dependencies, and removes the need for a GitHub token or network access for help output and local actions.

The user-facing contract:

```text
--help                 → help without contacting GitHub
diff a.json b.json     → reading local snapshots
snapshot 2021          → a service with the token of this very invocation
```

The scenario from [[issue-50]] has to be completed while preserving the settings precedence and the ability to verify commands independently.

The initial research suggested [Wire](https://github.com/google/wire) and [Fx](https://github.com/uber-go/fx). Choosing between them is not the goal of the issue: any clear solution is enough, as long as unused services are not initialized and the used ones receive the current configuration.
