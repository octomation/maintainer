---
code: MAIN-124
id: 446
database_id: 5143146542
node_id: I_kwDOE2M9Zc8AAAABMo4wLg
status: open
title: "fetch: treat an empty upstream as a successful no-op"
labels: ["scope: code","scope: test","type: bug","severity: major","impact: medium","effort: easy"]
url: https://github.com/octomation/maintainer/issues/446
created_at: 2026-08-13T15:42:23Z
updated_at: 2026-08-13T15:42:23Z
---

# fetch: treat an empty upstream as a successful no-op

**Motivation:** a Git repository may validly exist before its first push. Reconciliation should treat an accessible upstream with zero refs as an up-to-date no-op, not as a failed fetch.

**Details**

Clone already accepts an empty upstream and materialises a checkout with an origin. A later fetch of the same upstream is classified as a transport failure because the Git library reports the zero-ref condition separately from its usual up-to-date result. This makes every apply partial even though the repository is reachable and no work is required.

Required behaviour:

- fetching an accessible upstream with zero refs succeeds without changing the worktree or local refs;
- a checkout created from an empty upstream remains reconcilable on every run;
- once the upstream receives its first ref, a subsequent fetch discovers it normally;
- authentication, authorization, not-found and network failures remain errors;
- an otherwise successful apply reports zero errors and exits successfully.

**PoC**

From a tracked checkout whose upstream exists but has no refs:

```bash
$ git ls-remote origin
$ echo $?
0

$ maintainer fetch --apply
error: fetch owner/repository: remote repository is empty
Error: apply incomplete: 1 action(s) failed, 0 conflict(s) unresolved
$ echo $?
3
```

The regression test should use a temporary empty bare upstream and must not depend on a hosted repository.
