---
code:
id: I_kwDOE2M9Zc5JsYVD
databaseId: 1236370755
number: 40
url: https://github.com/octomation/maintainer/issues/40
title: "github: contribution: extend heat map by .From(), .To(), and .Range()"
labels:
  - "scope: code"
  - "scope: test"
milestone: "[[milestone-1]]"
state: CLOSED
stateReason: COMPLETED
createdAt: 2022-05-15T19:12:46Z
updatedAt: 2022-06-15T10:15:49Z
lastEditedAt:
closedAt: 2022-06-02T19:33:44Z
---

# github: contribution: extend heat map by .From(), .To(), and .Range()

Make the boundaries of the contributions calendar available to range operations, so that commands agree on the available period, build the viewing window, and handle an empty snapshot. The point is better integration with `time.Range`.

The contract, illustrated:

```text
The snapshot holds the dates 2021-01-01 … 2021-12-31.
Start: 2021-01-01; end: 2021-12-31.
An empty snapshot: no boundaries, and no arbitrary date is substituted.
```

The user-visible effect is more reliable lookup, diff and suggestion at the edges of the data. This task requires no CLI command of its own.

**Current state:** the issue is closed. [HeatMap](../../internal/model/github/contribution/heatmap.go) has `From()` and `To()`, and an empty map returns zero time values; the `.Range()` method from the title is absent. That is partial conformance to the originally requested API, not grounds for declaring the missing method implemented.

<!-- 2022-06-02T19:33Z https://github.com/octomation/maintainer/issues/40#issuecomment-1145259485
https://github.com/octomation/maintainer/commit/63ed3291dadb8e0d97f4c899b22716b9663adcb0
-->
