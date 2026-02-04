---
code: MAIN-131
id: I_kwDOE2M9Zc8AAAABYurjNA
databaseId: 5954528052
number: 497
url: https://github.com/octomation/maintainer/issues/497
title: "command: exit codes: enforce the documented codes for usage errors"
issueType: null
labels:
  - "scope: code"
  - "scope: test"
  - "type: bug"
  - "severity: minor"
  - "impact: medium"
  - "effort: easy"
assignees:
  - kamilsk
milestone: "[[MAIN-M2]]"
parent: null
issueFields: {}
state: CLOSED
stateReason: COMPLETED
duplicateOf: null
createdAt: 2026-10-21T04:48:56Z
updatedAt: 2026-10-23T05:46:48Z
lastEditedAt: null
closedAt: 2026-10-21T04:48:59Z
---

# command: exit codes: enforce the documented codes for usage errors

**Motivation:** the fetch and status guides promise `0/1/2/3`, and the JSON recipes rely on a script telling a usage error from a transport or Git failure. A mistyped subcommand must not succeed silently, and an invalid flag must exit with the documented code.

**Requirements**

- An unknown subcommand of any group is a usage error: a message in stderr naming the command and exit `2`; `main_test.go` asserts that `exit` is called with the code.
- A flag parse error and an invalid argument exit `2`; a negative `--concurrency` is rejected by both commands instead of being replaced by the config value; an unknown `--format` is rejected by `fetch` as it is by `status`, before any request, instead of falling back to the human output.
- The exit-code tables in `fetch.md` and `status.md` promise `2` for every usage error and stay true.

**Evidence:** findings 17 (`internal/command/root.go:20`, `main_test.go:21`) and 18 (`main.go:59`, `internal/command/fetch/root.go:104`, `docs/content/fetch.md:299`, `docs/content/status.md:217`) of [the code review](../reviews/20260926T071417Z.md), reproduced on `932b9c8`; the `--format` case comes from [the issue review of 2026-10-19](../reviews/20261019T105848Z.md#main-131-exit-codes-of-usage-errors).

**PoC**

```bash
maintainer fecth; echo "exit=$?"                    # help and 0 today — expected: "unknown command", 2
maintainer status --concurrency=abc; echo "exit=$?"  # 1 today — expected: 2
maintainer fetch --timeout=soon; echo "exit=$?"      # 1 today — expected: 2
maintainer fetch --concurrency=-1; echo "exit=$?"    # silently uses the config today — expected: rejected, 2
maintainer fetch --format=jsno; echo "exit=$?"       # human output today — expected: rejected before any request, 2
```

**Out of scope:** the stdout/stderr split of `lookup` and `histogram` (finding 30), deferred to a separate contribution fix with output-contract tests.

Related: [[MAIN-132]], [[MAIN-50]].
