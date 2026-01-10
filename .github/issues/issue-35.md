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
updatedAt: 2022-05-12T14:36:21Z
lastEditedAt:
closedAt: 2022-05-12T14:36:21Z
---

# define tool config

Make configuring maintainer predictable: command options and environment variables must land in one shared configuration with a clear precedence. This completes the token-passing scenario from [#33](issue-33.md) and reduces how much the commands depend on a hidden environment.

An example of the working scenario:

```bash
GITHUB_TOKEN="$PERSONAL_GITHUB_TOKEN" maintainer github contribution snapshot 2021
maintainer github --token="$WORK_GITHUB_TOKEN" contribution snapshot 2021
```

**Current state:** the issue is closed. Settings are read from a local `.env`, from the `GITHUB_TOKEN` / `GIT_REMOTE` variables, and from the `--token` / `--remote` flags. Loading from the home directory is still marked as further work, and there is no general `config` command. The fact that a remote setting exists does not mean every contribution command uses it to select the user.

The result is verified by the [configuration and its tests](../../internal/config/tool_test.go). Richer profiles and workspaces from the notes belong to the separate fetch specification.

<!-- 2022-05-12T14:36Z https://github.com/octomation/maintainer/issues/35#issuecomment-1125074780
fixed
-->
