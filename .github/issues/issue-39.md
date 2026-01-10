---
code:
id: I_kwDOE2M9Zc5JsYPz
databaseId: 1236370419
number: 39
url: https://github.com/octomation/maintainer/issues/39
title: "org: review and classify issues"
labels: []
milestone:
state: OPEN
stateReason:
createdAt: 2022-05-15T19:11:31Z
updatedAt: 2026-09-25T05:12:02Z
lastEditedAt: 2026-09-25T05:12:02Z
closedAt:
---

# org: review and classify issues

Bring the backlog to a state where it is clear what to do next and why. Terse titles, duplicates, and completed work mixed with live ideas all get in the way of assessing priorities.

The original actions: add meaningful prefixes, distribute the tasks across projects, close the unnecessary ones. Before changing any status, the code and the documents have to be checked: a closed issue may describe a removed feature, and an open one a partially implemented feature.

An example of the classification result:

```text
github: contribution: ... → a calendar defect, reproduction available
template: ...             → a maintenance improvement, needs a prototype
org: ...                  → an organizational outcome and release boundaries
```

Readiness means every task has an understandable motivation, an expected outcome, and its relations; duplicates explain where the work continues. External statuses and projects are changed by a separate action — editing Markdown locally does not update them.

Additional context on a centralized inbox and prioritization: [delayed automation](<../notes/Issues/delayed automation.md>).
