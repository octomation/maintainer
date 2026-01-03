---
id: 133
database_id: 1666111192
node_id: I_kwDOE2M9Zc5jTtbY
status: open
title: "github: contribution: suggest for cmm"
labels: ["scope: code","scope: test","type: bug","severity: major","impact: medium","effort: easy"]
url: https://github.com/octomation/maintainer/issues/133
created_at: 2023-04-13T10:02:01Z
updated_at: 2023-04-13T10:02:02Z
---

# github: contribution: suggest for cmm

Do not crash when the author date of HEAD lies in the future relative to the current time. This happens when a Git wrapper is used again after a commit with a future timestamp.

The original scenario:

```text
git contrib … → creates a commit dated later than the present moment
git contrib … → recovered: assertion is not a true
                unexpected panic occurred
```

The expected behaviour is to limit the reference moment by the current time, as the original report proposed, or else to state clearly that there is no valid suggestion. An impossible range must not be constructed, and an empty date must not be handed to the next command.

**Code context:** suggest clamps the end of the range to `now` and only then sets the start from the HEAD timestamp; with a future HEAD the start can end up later than the end. That is a verifiable cause of the panic risk, not an assumption about how `git contrib` — an external wrapper — behaves.

Cases needed: a future time today, and a future calendar date. The detailed reproduction with clock times and a stack is [#148](issue-148.md); an excessive random offset is [#193](issue-193.md).
