---
code:
id: I_kwDOE2M9Zc5psGu5
databaseId: 1773169593
number: 148
url: https://github.com/octomation/maintainer/issues/148
title: "github: contribution: edge case for contrib suggestion"
labels:
  - "scope: code"
  - "type: bug"
  - "severity: critical"
  - "impact: medium"
  - "effort: medium"
milestone: "[[milestone-1]]"
state: OPEN
stateReason:
createdAt: 2023-06-25T09:20:51Z
updatedAt: 2023-06-25T09:20:52Z
lastEditedAt:
closedAt:
---

# github: contribution: edge case for contrib suggestion

Handle a future HEAD date without a panic and without handing an empty timestamp to Git. This is a concrete reproduction of the defect class from [#133](issue-133.md), which additionally ended with `fatal: invalid date format:` in the external wrapper.

The recorded conditions:

```text
Current time: 2023-06-25 12:15:27 +0300
HEAD date:    2023-06-26 08:35:26 +0300
Command:      git contrib dev: add init task
Result:       assertion is not a true → invalid date format
```

The useful part of the stack: `Suggest → Range.Since → NewRange`. The selection tried to build a range from a future HEAD to the present moment.

The expectation is either clamping the anchor by an agreed rule, or an explicit error stating that no valid time exists; neither a zero nor an empty date may look like a successful suggestion. Cases to check: a future hour today, the next day, a future year, and an ordinary HEAD in the past.

**At present** the order in which the bounds are set in [suggest](../../internal/command/github/contribution/suggest.go) preserves this risk. `git contrib` is an external command; maintainer is responsible for correct stdout and a correct exit status of its own invocation.
