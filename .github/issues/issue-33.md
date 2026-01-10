---
code:
id: I_kwDOE2M9Zc5JfBkV
databaseId: 1232869653
number: 33
url: https://github.com/octomation/maintainer/issues/33
title: "provide github access token by command flag"
labels:
  - "scope: code"
milestone: "[[milestone-1]]"
state: CLOSED
stateReason: COMPLETED
createdAt: 2022-05-11T15:55:41Z
updatedAt: 2022-06-15T10:15:49Z
lastEditedAt:
closedAt: 2022-05-12T20:14:10Z
---

# provide github access token by command flag

Allow a GitHub access token to be passed as a flag for a single invocation. That is convenient when working with several accounts, and in scripts where changing the environment of the whole shell session is undesirable.

The working example:

```bash
maintainer github --token="$WORK_GITHUB_TOKEN" contribution snapshot 2021
```

The alternative is the `GITHUB_TOKEN` variable. An explicit flag must take precedence for that run; invocations without it keep using the environment configuration.

**Current state:** the task is closed, `--token` is registered on the `github` group, and the flag and environment bindings are covered by the [configuration tests](../../internal/config/tool_test.go). The example also appears in [the documentation](../../docs/changelog.md). The overall tool configuration is described by [#35](issue-35.md).
