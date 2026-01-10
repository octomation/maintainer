---
code:
id: I_kwDOE2M9Zc5JeYHg
databaseId: 1232699872
number: 31
url: https://github.com/octomation/maintainer/issues/31
title: "command: github backup repository"
labels: []
milestone:
state: OPEN
stateReason:
createdAt: 2022-05-11T14:06:20Z
updatedAt: 2023-01-06T13:48:42Z
lastEditedAt:
closedAt:
---

# command: github backup repository

Create a backup of a GitHub repository together with its maintenance data. A plain clone preserves the Git history but does not replace an archive of issues, labels and projects, without which the development context is lost.

A provisional interface:

```bash
maintainer github backup repository octomation/maintainer
```

A prototype of the result:

```text
backup/octomation/maintainer/
  repository.git/
  issues.json
  labels.json
  projects.json
  manifest.json
```

The full contents of the copy have to be defined, preserving the relations between objects, their identifiers and the moment of the dump. A partial failure must not look like a successful full backup; the result must make it possible to verify what was actually stored. Restoring requires a separate contract.

**At present** there is no backup command; the existing draft for reading issues is not wired into the CLI and is not an archiver. The [fetch specification](<../notes/Specs/GitHub fetcher, PoC implementation plan.md>) explicitly excludes downloading issues and projects, and excludes mirroring. Related context: [Repository Metadata](<../notes/Repository Metadata.md>) and [raw dumps](<../notes/Issues/maintainer as raw.md>).
