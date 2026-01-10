---
code:
id: I_kwDOE2M9Zc5JiT8I
databaseId: 1233731336
number: 35
url: https://github.com/octomation/maintainer/issues/35
title: "define tool config"
labels:
  - "scope: code"
  - "scope: test"
milestone:
state: CLOSED
stateReason: COMPLETED
createdAt: 2022-05-12T09:45:18Z
updatedAt: 2026-09-25T05:11:52Z
lastEditedAt: 2026-09-25T05:11:52Z
closedAt: 2022-05-12T14:36:21Z
---

# define tool config

Make configuring maintainer predictable: command options and environment variables must land in one shared configuration with a clear precedence. This completes the token-passing scenario from [[issue-33]] and reduces how much the commands depend on a hidden environment.

An example of the working scenario:

```bash
GITHUB_TOKEN="$PERSONAL_GITHUB_TOKEN" maintainer github contribution snapshot 2021
maintainer github --token="$WORK_GITHUB_TOKEN" contribution snapshot 2021
```

Richer profiles and workspaces from the notes belong to the separate fetch specification.

<!-- 2022-05-12T14:36Z https://github.com/octomation/maintainer/issues/35#issuecomment-1125074780
fixed
-->
