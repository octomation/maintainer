---
code:
id: I_kwDOE2M9Zc5auZU5
databaseId: 1522111801
number: 90
url: https://github.com/octomation/maintainer/issues/90
title: "github: contribution: lookup call throw panic"
labels:
  - "scope: code"
milestone: "[[milestone-1]]"
state: CLOSED
stateReason: COMPLETED
createdAt: 2023-01-06T07:32:55Z
updatedAt: 2023-01-06T13:48:01Z
lastEditedAt:
closedAt: 2023-01-06T13:48:00Z
---

# github: contribution: lookup call throw panic

Handle changes to the GitHub calendar format without an unhandled panic. In the original case lookup terminated with `panic: invalid count value:`, so the user received neither a calendar nor a clear explanation of the source failure.

The reproduction from the report:

```bash
maintainer github contribution lookup /-2
# panic: invalid count value:
```

The stack pointed at reading the count out of the HTML. Either correct recognition of the available format, or a diagnosable load/parse error, is expected; an empty value must not silently turn the whole calendar into zeros.

**Current state:** this historical issue is closed and the parser has changed since. The current code still panics on some malformed data — `invalid count value` is still raised from the heat-map builder — so closure does not amount to general protection against any future GitHub change. Later incidents: the move to a table [#150](issue-150.md), the tooltip [#174](issue-174.md), asynchronous loading [#220](issue-220.md).
