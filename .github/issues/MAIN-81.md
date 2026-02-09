---
code: MAIN-81
id: I_kwDOE2M9Zc5PshxL
databaseId: 1337072715
number: 81
url: https://github.com/octomation/maintainer/issues/81
title: "makefile: build: allow include recursively"
issueType: null
labels:
  - "scope: code"
  - "scope: test"
assignees:
  - kamilsk
milestone: "[[MAIN-M1]]"
parent: null
issueFields: {}
state: OPEN
stateReason: null
duplicateOf: null
createdAt: 2022-08-12T11:11:50Z
updatedAt: 2026-11-21T05:18:09Z
lastEditedAt: 2026-11-21T05:18:09Z
closedAt: null
---

# makefile: build: allow include recursively

## Motivation

`maintainer makefile build` assembled one self-contained Makefile from the
fragments of octomation/makefiles by expanding their `include` lines. A fragment
could not include what it needs: `build.service.mk` uses `build.tool.mk`, but
the relation was declared only at the top, in `Go Service.mk`. Every top-level
file that took the service fragment had to repeat the tool fragment.

The goal: a fragment includes its own dependencies, and the assembly expands
them at any depth. The author added in 2023 that recursion and deep nesting
must be guarded, so a cycle cannot hang the build.

## PoC

```make
# Go Service.mk
include build.service.mk

# build.service.mk
include build.tool.mk
```

```bash
maintainer makefile build 'Go Service.mk'
```

## Acceptance criteria

- [ ] The chain `Go Service.mk` → `build.service.mk` → `build.tool.mk`
      assembles into one file, with the rules in include order.
- [ ] An include path resolves against the directory of the including file, not
      the working directory.
- [ ] A fragment included twice appears once.
- [ ] A missing `-include` is skipped; a missing `include` is an error that
      names the chain.
- [ ] A cycle such as `a.mk` ↔ `b.mk` is an error that names it, not a hang or
      `too many open files`.

Closing this means checking the original fragment structure and these cases,
not only that the call is recursive.

## Out of scope

- The repository's own `Makefile`, sourced from octomation/makefiles.

Related: [[MAIN-162]].

<!-- 2023-08-09T12:38Z https://github.com/octomation/maintainer/issues/81#issuecomment-1671248449
I have to avoid recursive including and deep-nesting cases.
-->

<!-- 2026-11-20T21:00Z https://github.com/octomation/maintainer/issues/81#issuecomment-7043719750
Maturity: partly built, then removed. `Makefile.AppendTo` in `internal/command/makefile/entity.go` expanded `include` and `-include` recursively since the 2022 snapshot; paths resolved against the working directory and there was no visited set, so `a.mk` <-> `b.mk` failed with `too many open files` (reviews of 2026-09-13, 2026-09-25 and 2026-09-26).
The command and `internal/command/makefile` went in commit 7cc45e0 (MAIN-162).
v0.5.0 proposal: not planned. Assembling Makefiles is a build helper, not work with repositories, issue trackers or GitHub; octomation/makefiles does not call it.
-->
