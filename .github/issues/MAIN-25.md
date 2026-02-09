---
code: MAIN-25
id: I_kwDOE2M9Zc5JKbHl
databaseId: 1227469285
number: 25
url: https://github.com/octomation/maintainer/issues/25
title: "add possibility to compare configuration of repos"
issueType: null
labels: []
assignees:
  - kamilsk
milestone: "[[MAIN-M3]]"
parent: null
issueFields: {}
state: OPEN
stateReason: null
duplicateOf: null
createdAt: 2022-05-06T06:26:52Z
updatedAt: 2026-11-20T15:42:35Z
lastEditedAt: 2026-11-20T15:42:35Z
closedAt: null
---

# add possibility to compare configuration of repos

## Motivation

A family of projects should be configured the same way: the same features
enabled, the same merge settings, the same protections. Today a difference is
found by opening the settings pages of each repository on GitHub one by one,
so drift goes unnoticed until it bites. The idea is to compare the settings of
repositories against a chosen reference and show the deviations, and later to
apply the reference to the others.

## PoC

```bash
$ maintainer config compare octomation/maintainer kamilsk/retry
# reference: octomation/maintainer
# kamilsk/retry
#   issues:         enabled → disabled
#   merge settings: squash only → merge, squash, rebase
#   unreadable:     branch protection (no admin access)
```

Applying the difference was proposed as an **optional follow-up**:

```bash
$ maintainer config apply octomation/maintainer kamilsk/retry kamilsk/semaphore
```

## Acceptance criteria

- [ ] `compare` changes nothing on GitHub.
- [ ] The set of compared settings is defined explicitly and documented.
- [ ] The output names the reference, every property that differs, and every
  property that could not be read; lack of access is never shown as an absent
  or disabled setting.
- [ ] A JSON mode gives the same comparison for further processing.
- [ ] If `apply` is built, it rests on a reviewable plan of transitions
  ([[MAIN-19]]) and changes only what the plan lists.

## Out of scope

- Synchronizing template files: [[MAIN-10]].
- Changing GitHub settings from `fetch`: fetch reads repositories, it does not
  configure them.

Related: [[MAIN-19]], [[MAIN-10]].

<!-- 2026-11-20T21:00Z https://github.com/octomation/maintainer/issues/25#issuecomment-7043716345
Maturity: idea. There is no config command group (internal/command/root.go). The closest code is fetch discovery, which reads a repository's identity, visibility, default branch and the fork, template and archived flags (internal/service/github/discovery_rest.go), but no feature or merge settings.
Proposal for v0.5.0: keep. Comparing the configuration of the contributor's repositories on GitHub is repository work across all projects, the one picture the mission asks for, and the read-only compare can reuse the GitHub client of fetch. Not a v0.5.0 deliverable: it stays in its milestone.
-->
