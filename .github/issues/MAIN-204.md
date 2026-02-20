---
code: MAIN-204
id: I_kwDOE2M9Zc8AAAABjEV1zw
databaseId: 6648329679
number: 582
url: https://github.com/octomation/maintainer/issues/582
title: "install: install from install.octolab.org/maintainer"
issueType: Task
labels:
  - "effort: easy"
  - "impact: low"
  - "scope: docs"
assignees:
  - kamilsk
milestone: null
parent: null
issueFields: {}
state: OPEN
stateReason: null
duplicateOf: null
createdAt: 2026-11-27T14:24:54Z
updatedAt: 2026-11-27T15:02:03Z
lastEditedAt: 2026-11-27T15:02:03Z
closedAt: null
---

# install: install from install.octolab.org/maintainer

## Motivation

The install script lives in this repository as `bin/install`, a copy that
godownloader made in 2023 and nothing regenerates. Since octomation/install-action,
every stable release renders its own installer from the release and writes
it to `https://install.octolab.org/maintainer`.

## Scope

- README and the quick start install from `install.octolab.org/maintainer`.
- The release checks install the tag through that installer.
- The doctor proves that the Releaser App can write to `octolab/install`.

## Acceptance criteria

- [ ] README and `docs/content/quickstart.md` name only
      `https://install.octolab.org/maintainer`.
- [ ] After the release, the header of `https://install.octolab.org/maintainer`
      names install-action, `sh -s -- -h` prints the usage and the platforms
      of the release, and the script installs the tag.
- [ ] `doctor.yml` mints a token for `octolab/install`.

## Out of scope

Removing `bin/install`: published release notes link it.
