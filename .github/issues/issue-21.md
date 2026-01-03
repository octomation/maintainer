---
id: 21
database_id: 1034379376
node_id: I_kwDOE2M9Zc49p2Bw
status: open
title: "improve vanity support"
labels: ["scope: code"]
url: https://github.com/octomation/maintainer/issues/21
created_at: 2021-10-24T10:30:31Z
updated_at: 2023-01-06T13:48:51Z
---

# improve vanity support

Extend vanity generation to repositories with several Go modules and with import paths that do not match their directories. The build reads a hand-written `modules.yml` today, which is not enough to maintain a complex project layout comfortably.

The original cases to spell out and support:

- Dumping the module description (`dump`) so the data is visible before pages are generated.
- Several modules inside one repository.
- A directory mapped to a different public path: `reporters/<reporter>` → `go.qase.io/reporter/<reporter>`.
- Linking a file through `blob` and a directory through `tree`, instead of using `tree` for both.

The criterion of success: the import path of every module resolves to the right repository, and source links lead to the intended file or directory. Subpackages without a root package are already covered by [#6](issue-6.md); automatic discovery of sources is [#134](issue-134.md).

The original qase-go examples: [dump](https://github.com/qase-tms/qase-go/commit/ffaccae028365f842cabc0bd63cb071171d2b47e) and [submodules](https://github.com/qase-tms/qase-go/commit/e12c8ca0c17082970e1af13bd11efb56d0408dfe).
