---
code:
id: MDU6SXNzdWU4NDEwNzUyNTc=
databaseId: 841075257
number: 15
url: https://github.com/octomation/maintainer/issues/15
title: "extend octolab preset by other repositories"
labels:
  - "scope: code"
milestone:
state: CLOSED
stateReason: NOT_PLANNED
createdAt: 2021-03-25T15:59:21Z
updatedAt: 2026-09-25T05:10:30Z
lastEditedAt: 2026-09-25T05:10:30Z
closedAt: 2023-03-31T15:33:04Z
---

# extend octolab preset by other repositories

Extend the OctoLab preset to repositories with different purposes, and allow label sets to be combined. A single set built for code does not always suit documentation, RFCs or workshop material.

The original scope:

- New subset: `kamilsk/workshops`, `octolab/docs`, `octolab/rfc`.
- Empty set, meaning label removal: `kamilsk/gex`.
- Update: `kamilsk/bridge` (from defaults), plus `breaker`, `check`, `dotfiles`, `egg`, `genome`, `grafaman` under the same owner.

The historical prototypes:

```bash
maintainer github labels patch octolab hacktoberfest
maintainer github labels patch empty
```

The task was to define how presets merge, to show conflicting definitions, and to distinguish adding a thematic set of labels from removing a set entirely.

Related: presets [[issue-13]], label scope removal [[issue-80]].
