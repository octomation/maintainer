---
id: 15
database_id: 841075257
node_id: MDU6SXNzdWU4NDEwNzUyNTc=
status: closed
title: "extend octolab preset by other repositories"
labels: ["scope: code"]
url: https://github.com/octomation/maintainer/issues/15
created_at: 2021-03-25T15:59:21Z
updated_at: 2023-03-31T15:33:04Z
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

**Current state:** a closed follow-up to [#13](issue-13.md). After labels were handed over to the settings app ([#80](issue-80.md)), these commands do not exist in maintainer.
