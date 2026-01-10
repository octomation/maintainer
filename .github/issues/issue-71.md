---
code:
id: I_kwDOE2M9Zc5ObHko
databaseId: 1315731752
number: 71
url: https://github.com/octomation/maintainer/issues/71
title: "pkg: file: register encoders for specific formats"
labels:
  - "scope: code"
  - "scope: test"
milestone: "[[milestone-1]]"
state: CLOSED
stateReason: COMPLETED
createdAt: 2022-07-23T19:38:09Z
updatedAt: 2022-07-25T19:29:56Z
lastEditedAt:
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

**Current state:** the issue is closed; [Packer](../../internal/pkg/io/packer.go) is used by the contributions file source and is covered by tests. `diff` can read JSON and YAML files, but `snapshot` writes JSON to stdout: the existence of a shared mechanism does not give it a `--format` flag it never had. The previous helper area was removed in [#74](issue-74.md).
