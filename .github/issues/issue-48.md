---
code:
id: I_kwDOE2M9Zc5LKhhd
databaseId: 1261049949
number: 48
url: https://github.com/octomation/maintainer/issues/48
title: "ci/cd: tools: test act for local development"
labels: []
milestone:
state: OPEN
stateReason:
createdAt: 2022-06-05T13:29:43Z
updatedAt: 2023-01-06T13:48:49Z
lastEditedAt:
closedAt:
---

# ci/cd: tools: test act for local development

Check how much running GitHub Actions locally speeds up work on maintainer's workflows. The verification cycle depends on remote runs today; a local experience should give a fast way to investigate failures before changes are pushed.

The candidate is [act](https://github.com/nektos/act). A prototype of the experiment against the current workflow:

```bash
act -W .github/workflows/ci.yml -j test
```

The outcome of the task is a reproducible local-run scenario documenting the prerequisites, the jobs that were verified, and the differences from a hosted runner. A representative build/test check has to be chosen, and the external integrations that a local run does not cover have to be recorded.

**Current state:** there is no act-specific integration in the repository. This is research into a development tool, not a new maintainer command. The general idea of configuring and enabling workflows is described in [work with workflows](<../notes/Issues/work with workflows.md>).
