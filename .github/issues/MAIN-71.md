---
code: MAIN-71
id: I_kwDOE2M9Zc5ObHko
databaseId: 1315731752
number: 71
url: https://github.com/octomation/maintainer/issues/71
title: "pkg: file: register encoders for specific formats"
issueType: null
labels:
  - "scope: code"
  - "scope: test"
assignees:
  - kamilsk
milestone: "[[MAIN-M1]]"
parent: null
issueFields: {}
state: CLOSED
stateReason: COMPLETED
duplicateOf: null
createdAt: 2022-07-23T19:38:09Z
updatedAt: 2026-09-25T05:13:29Z
lastEditedAt: 2026-09-25T05:13:29Z
closedAt: 2022-07-25T19:29:56Z
---

# pkg: file: register encoders for specific formats

Make reading and writing snapshots uniform, driven by the file format. The user must not meet different format rules in different commands, and adding a new representation must not require repeating the same switch everywhere.

The contract, illustrated:

```text
snapshot.json → JSON
snapshot.yml  → YAML
snapshot.yaml → YAML
snapshot.txt  → a clear unsupported-format error
```

The criterion of success: identical data after a write followed by a read, and correct messages for corrupted content and for an unknown extension.

Related: removing the previous helper area [[MAIN-74]].
