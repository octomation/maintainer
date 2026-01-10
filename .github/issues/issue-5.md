---
code:
id: MDU6SXNzdWU3ODAzNzI0NTg=
databaseId: 780372458
number: 5
url: https://github.com/octomation/maintainer/issues/5
title: "integrate or implement embedmd"
labels: []
milestone:
state: OPEN
stateReason:
createdAt: 2021-01-06T09:33:18Z
updatedAt: 2023-01-06T13:48:57Z
lastEditedAt:
closedAt:
---

# integrate or implement embedmd

Keep Markdown examples in sync with the source files they come from. Examples in the README and under `docs/` are maintained by hand today, so a code change easily leaves the documentation holding a stale copy.

The reference is [embedmd](https://github.com/campoy/embedmd): the document stores a reference to a file or a fragment of it, and the tool refreshes the code block that follows. A prototype of the markup for an existing file:

```markdown
[embedmd]:# (main.go go /func main/ $)
```

Either integrating the existing tool or implementing compatible behaviour in maintainer is acceptable. The user must be able to refresh the insertions and to see divergences without modifying the file; a bad path or fragment boundary must produce a clear error. Text around the insertions is preserved, and a re-run without changes produces no diff.

**Current state:** no such command exists. Related: Markdown processing [#11](issue-11.md) and the idea of generating documents from a single source in [Repository Metadata](<../notes/Repository Metadata.md>).
