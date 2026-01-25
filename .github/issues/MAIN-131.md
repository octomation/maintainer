---
code: MAIN-131
id: null
databaseId: null
number: null
url: null
title: "command: exit codes: enforce the documented codes for usage errors"
labels:
  - "type: bug"
  - "scope: code"
  - "scope: test"
  - "severity: minor"
  - "impact: medium"
  - "effort: easy"
milestone: "[[MAIN-M3]]"
state: OPEN
stateReason: null
createdAt: 2026-09-26T09:31:00Z
updatedAt: 2026-09-26T09:31:00Z
lastEditedAt: null
closedAt: null
---

# command: exit codes: enforce the documented codes for usage errors

**Motivation:** the fetch and status guides promise `0/1/2/3`, and the JSON recipes rely on a script telling a usage error from a transport or Git failure. A mistyped subcommand must not succeed silently, and an invalid flag must exit with the documented code.

**Requirements**

- An unknown subcommand of any group is a usage error: a message in stderr naming the command and exit `2`; `main_test.go` asserts that `exit` is called with the code.
- A flag parse error and an invalid argument exit `2`; a negative `--concurrency` is rejected by both commands instead of being replaced by the config value.
- The exit-code tables in `fetch.md` and `status.md` stay true.

**Evidence:** findings 17 (`internal/command/root.go:20`, `main_test.go:21`) and 18 (`main.go:59`, `internal/command/fetch/root.go:104`, `docs/content/fetch.md:299`, `docs/content/status.md:217`) of [the code review](../reviews/20260926T071417Z.md), reproduced on `932b9c8`.

**PoC**

```bash
maintainer fecth; echo "exit=$?"                    # help and 0 today — expected: "unknown command", 2
maintainer status --concurrency=abc; echo "exit=$?"  # 1 today — expected: 2
maintainer fetch --timeout=soon; echo "exit=$?"      # 1 today — expected: 2
maintainer fetch --concurrency=-1; echo "exit=$?"    # silently uses the config today — expected: rejected, 2
```

**Out of scope:** the stdout/stderr split of `lookup` and `histogram` (finding 30), deferred to a separate contribution fix with output-contract tests.

Related: [[MAIN-132]], [[MAIN-50]].
