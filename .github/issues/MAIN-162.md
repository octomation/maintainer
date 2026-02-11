---
code: MAIN-162
id: I_kwDOE2M9Zc8AAAABhOjwLA
databaseId: 6524825644
number: 537
url: https://github.com/octomation/maintainer/issues/537
title: "makefile: remove the makefile build command"
issueType: Task
labels:
  - "scope: docs"
  - "scope: code"
  - "type: improvement"
assignees:
  - kamilsk
milestone: null
parent: null
issueFields: {}
state: CLOSED
stateReason: COMPLETED
duplicateOf: null
createdAt: 2026-11-21T05:16:42Z
updatedAt: 2026-11-21T06:11:56Z
lastEditedAt: 2026-11-21T06:11:56Z
closedAt: 2026-11-21T06:05:53Z
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
- [x] The changelog of v0.5.0 names the removal.
- [x] [[MAIN-81]] is closed as not planned with a link here.

## Out of scope

- The repository's own `Makefile`, sourced from octomation/makefiles.
- Changelogs of past releases.

<!-- 2026-11-20T20:00Z https://github.com/octomation/maintainer/issues/537#issuecomment-7057807416
Prepared in the working tree for v0.5.0; local only, not on GitHub yet.

- `internal/command/makefile` and `docs/content/makefiles.md` are gone, `root.go` no longer wires the command; `go build` and `go test ./...` pass, `go.mod` is unchanged.
- The home page loses the «ALSO IN THE BOX» block and its CSS; the release band takes its 120px (80px on narrow screens) gap so the spacing after the contributions section stays.
- `references/cli.md` is regenerated with `go generate ./skills`; `next build` passes.
-->
