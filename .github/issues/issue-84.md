---
code:
id: I_kwDOE2M9Zc5RP1hW
databaseId: 1363105878
number: 84
url: https://github.com/octomation/maintainer/issues/84
title: "github: contribution: invalid suggestion for past"
labels:
  - "scope: code"
  - "scope: test"
  - "type: bug"
milestone: "[[milestone-1]]"
state: CLOSED
stateReason: COMPLETED
createdAt: 2022-09-06T11:01:52Z
updatedAt: 2026-09-25T05:14:03Z
lastEditedAt: 2026-09-25T05:14:03Z
closedAt: 2023-03-25T20:26:24Z
---

# github: contribution: invalid suggestion for past

Do not propose a date that breaks the chosen commit chronology. In the original scenario the suggestion returned 11 November 2021, although the history under consideration already held a commit dated 12 November; after the wrapper was used, HEAD ended up dated earlier than its parent.

The original call:

```bash
maintainer github contribution suggest --delta 2021/+5
# Obtained: 2021-11-11, 7 → 9
```

The expected principle: the search respects an agreed lower bound and does not roll time backwards. An explicitly given period must be distinguished from an anchor on the Git history: the bare word `2021` says nothing about the date of the latest commit.

Verification has to pin both the input date and the history, rather than declaring any choice in November wrong.

Related: passing a timestamp directly [[issue-76]], a choice before a specific date [[issue-119]].
