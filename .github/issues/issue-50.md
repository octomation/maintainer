---
code:
id: I_kwDOE2M9Zc5LUxM7
databaseId: 1263735611
number: 50
url: https://github.com/octomation/maintainer/issues/50
title: "command: separate command structure and command execution"
labels:
  - "scope: code"
  - "scope: test"
  - "type: improvement"
  - "impact: medium"
  - "effort: medium"
milestone: "[[milestone-1]]"
state: OPEN
stateReason:
createdAt: 2022-06-07T18:59:04Z
updatedAt: 2023-04-06T06:13:34Z
lastEditedAt:
closedAt:
---

# command: separate command structure and command execution

Make the commands verifiable without real GitHub requests and without depending on the developer's environment. The configuration only appears at run time, so testing authorization, file and date failures is hard while constructing a command is bound to concrete services.

An example of a check that is needed: for the same arguments, substitute a successful calendar response, a network refusal, and a corrupted snapshot; then assert the output and the error of each scenario. The user must get identical behaviour regardless of how the command was constructed.

**Current state:** the registration and execution files are already partly separated, but the contribution commands still construct the GitHub service inside the run. The original motivation stands; the particular runner signature from the initial PoC is not a requirement.

The completion criterion: dependencies can be substituted when a command is tested, configuration is loaded in time, and the help needs no network access. Lazy service resolution is the related task [#51](issue-51.md).
