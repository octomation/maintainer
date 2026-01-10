---
code:
id: I_kwDOE2M9Zc5L-pm7
databaseId: 1274714555
number: 60
url: https://github.com/octomation/maintainer/issues/60
title: "github: contribution: replace --weeks by /weeks argument format for suggest"
labels:
  - "scope: docs"
  - "scope: code"
milestone: "[[milestone-1]]"
state: CLOSED
stateReason: COMPLETED
createdAt: 2022-06-17T08:08:26Z
updatedAt: 2022-07-22T19:57:08Z
lastEditedAt:
closedAt: 2022-07-22T19:57:08Z
---

# github: contribution: replace --weeks by /weeks argument format for suggest

Unify how the period is given to suggest and to lookup. A single format reduces the number of flags and lets the chosen viewing window be carried into the date-selection command.

The working forms:

```bash
maintainer github contribution suggest 2013-11-20/5
maintainer github contribution suggest 2013-11-20/+2
maintainer github contribution suggest 2013-11-20/-2
```

Without a sign, a window around the date is requested; `+` and `-` set the direction relative to the reference week. Note that suggest still never picks a date earlier than the given moment: seeing an earlier part of the calendar does not by itself license advising an earlier commit.

**Current state:** the issue is closed, the argument is parsed jointly with lookup, and there is no `--weeks` flag. The original point of comparison is the [historical interface](https://github.com/octomation/maintainer/blob/1a68d3ea715fd5b22500dc7a1c081fcca95784ad/internal/command/github/contribution.go#L188-L205). The semantics of weeks are covered in [#41](issue-41.md), the regressions in [#72](issue-72.md) and [#73](issue-73.md).
