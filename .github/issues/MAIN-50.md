---
code: MAIN-50
id: I_kwDOE2M9Zc5LUxM7
databaseId: 1263735611
number: 50
url: https://github.com/octomation/maintainer/issues/50
title: "command: separate command structure and command execution"
issueType: null
labels:
  - "scope: code"
  - "scope: test"
  - "type: improvement"
  - "impact: medium"
  - "effort: medium"
assignees:
  - kamilsk
milestone: "[[MAIN-M1]]"
parent: null
issueFields: {}
state: OPEN
stateReason: null
duplicateOf: null
createdAt: 2022-06-07T18:59:04Z
updatedAt: 2026-09-25T05:12:33Z
lastEditedAt: 2026-09-25T05:12:33Z
closedAt: null
---

# command: separate command structure and command execution

Make the commands verifiable without real GitHub requests and without depending on the developer's environment. The configuration only appears at run time, so testing authorization, file and date failures is hard while constructing a command is bound to concrete services.

An example of a check that is needed: for the same arguments, substitute a successful calendar response, a network refusal, and a corrupted snapshot; then assert the output and the error of each scenario. The user must get identical behaviour regardless of how the command was constructed.

The completion criterion: dependencies can be substituted when a command is tested, configuration is loaded in time, and the help needs no network access. The particular runner signature from the initial PoC is not a requirement. Lazy service resolution is the related task [[MAIN-51]].
