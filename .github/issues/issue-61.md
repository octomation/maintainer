---
code:
id: I_kwDOE2M9Zc5L_Jlt
databaseId: 1274845549
number: 61
url: https://github.com/octomation/maintainer/issues/61
title: "text: replace -> by →"
labels:
  - "scope: docs"
  - "scope: code"
milestone:
state: CLOSED
stateReason: COMPLETED
createdAt: 2022-06-17T10:12:05Z
updatedAt: 2022-06-22T06:39:18Z
lastEditedAt:
closedAt: 2022-06-22T06:39:18Z
---

# text: replace -> by →

Make textual transition markers uniform: use `→` instead of `->` in human-facing messages, so a "before → after" comparison reads the same way in suggestion and diff output. The motivation is clean, consistent text.

Example:

```text
Before: Suggestion is ..., 3 -> 5
After:  Suggestion is ..., 3 → 5
```

The change applies to the wording of messages and examples, not to shell syntax or the JSON format. Numbers and the direction of the transition keep their previous meaning.

**Current state:** the issue is closed; the current suggest and diff messages use `→`. The [older documentation](../../docs/changelog.md) still contains `0 -> 5`, so the closed status does not mean the symbol has already been replaced in every historical example.
