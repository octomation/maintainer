---
id: 21
database_id: 1034379376
node_id: I_kwDOE2M9Zc49p2Bw
status: open
title: "improve vanity support"
labels: [scope: code]
url: https://github.com/octomation/maintainer/issues/21
created_at: 2021-10-24T10:30:31Z
updated_at: 2023-01-06T13:48:51Z
---

# improve vanity support

- case 1: dump

https://github.com/qase-tms/qase-go/commit/ffaccae028365f842cabc0bd63cb071171d2b47e

- case 2: submodules

https://github.com/qase-tms/qase-go/commit/e12c8ca0c17082970e1af13bd11efb56d0408dfe

- case 3: submodules with different path

```
./reporters/<reporter> -> go.qase.io/reporter/<reporter>
```

- case 4: blob instead tree
