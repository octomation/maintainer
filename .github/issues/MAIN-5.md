---
code: MAIN-5
id: MDU6SXNzdWU3ODAzNzI0NTg=
databaseId: 780372458
number: 5
url: https://github.com/octomation/maintainer/issues/5
title: "integrate or implement embedmd"
issueType: null
labels: []
assignees:
  - kamilsk
milestone: "[[MAIN-M1]]"
parent: null
issueFields: {}
state: CLOSED
stateReason: NOT_PLANNED
duplicateOf: null
createdAt: 2021-01-06T09:33:18Z
updatedAt: 2026-11-21T05:33:37Z
lastEditedAt: 2026-11-20T15:42:40Z
closedAt: 2026-11-21T05:33:37Z
---

# integrate or implement embedmd

## Motivation

The examples in the README and under `docs/` are copied from source files and
command output by hand. A code change easily leaves the documentation with a
stale copy, and nothing tells.

The reference is [embedmd](https://github.com/campoy/embedmd): the document keeps
a reference to a file or a fragment of it, and the tool refreshes the code block
that follows. Either integrating the existing tool or implementing compatible
behaviour in maintainer is acceptable.

## PoC

The markup for an existing file:

```markdown
[embedmd]:# (main.go go /func main/ $)
```

The command name is provisional:

```bash
maintainer embedmd -w README.md    # refresh the code blocks in place
maintainer embedmd -d README.md    # print the divergences, change nothing
```

## Acceptance criteria

- [ ] The user can refresh the insertions, and can see the divergences without
      modifying the file; a divergence gives a non-zero exit, so CI can check it.
- [ ] A bad path or fragment boundary is an error that names the document line.
- [ ] Text around the insertions is preserved.
- [ ] A re-run with no source change produces no diff.

## Out of scope

- Parsing Markdown in general: [[MAIN-11]].
- Inserting command output, not file contents.

<!-- 2026-11-20T21:00Z https://github.com/octomation/maintainer/issues/5#issuecomment-7043718224
Maturity: idea. No embedding command is wired in `internal/command/root.go`, and no Markdown tooling is in `go.mod`. The part of the drift that hurt most, the CLI reference of the skill, is already generated from the commands (`go generate ./skills` writes `skills/maintainer/references/cli.md`).
v0.5.0 proposal: not planned. Keeping Markdown examples in sync is a documentation build helper, off the mission; campoy/embedmd already does it as a standalone tool any repository can run.
-->

<!-- 2026-11-21T05:30Z https://github.com/octomation/maintainer/issues/5#issuecomment-7057978818
Closed as not planned: this is no longer the focus of maintainer. From v0.5.0 maintainer is the contributor's central tool: one picture of the repositories, issue trackers and GitHub across all projects, not build or release tooling.
embedmd works on its own. Tooling for a new project's docs is to be thought through under #153, which stays open.
-->
