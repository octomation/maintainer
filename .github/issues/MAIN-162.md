---
code: MAIN-162
id: null
databaseId: null
number: null
url: null
title: "makefile: remove the makefile build command"
issueType: Task
labels:
  - "scope: code"
  - "scope: docs"
  - "type: improvement"
assignees:
  - kamilsk
milestone: null
parent: null
issueFields: {}
state: OPEN
stateReason: null
duplicateOf: null
createdAt: null
updatedAt: null
lastEditedAt: null
closedAt: null
---

# makefile: remove the makefile build command

## Motivation

v0.5.0 slims maintainer down to working with repositories and GitHub: fetch, status,
the contribution commands and the local hub. `makefile build`, which expands the
`include` lines of Makefiles, is a build helper of another kind, like `go vanity`
removed in [[MAIN-161]]; octomation/makefiles does not call it.

## PoC

```bash
maintainer makefile build --help
grep -rln makefile internal docs/content skills/maintainer
```

## Acceptance criteria

- [x] `maintainer makefile` is gone with `internal/command/makefile`; `go test ./...` passes.
- [x] The site has no `makefiles` page, menu item or quick start link; the home page has no
      «ALSO IN THE BOX» block; `next build` passes.
- [x] `docs/README.md` and the skill (`SKILL.md`, `references/workflows.md`,
      `references/cli.md`) no longer mention the command.
- [ ] The changelog of v0.5.0 names the removal.
- [ ] [[MAIN-81]] is closed as not planned with a link here.

## Out of scope

- The repository's own `Makefile`, sourced from octomation/makefiles.
- Changelogs of past releases.

<!-- 2026-11-20T20:00Z
Prepared in the working tree for v0.5.0; local only, not on GitHub yet.

- `internal/command/makefile` and `docs/content/makefiles.md` are gone, `root.go` no longer wires the command; `go build` and `go test ./...` pass, `go.mod` is unchanged.
- The home page loses the «ALSO IN THE BOX» block and its CSS; the release band takes its 120px (80px on narrow screens) gap so the spacing after the contributions section stays.
- `references/cli.md` is regenerated with `go generate ./skills`; `next build` passes.
-->
