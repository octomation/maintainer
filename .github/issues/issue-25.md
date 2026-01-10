---
code:
id: I_kwDOE2M9Zc5JKbHl
databaseId: 1227469285
number: 25
url: https://github.com/octomation/maintainer/issues/25
title: "add possibility to compare configuration of repos"
labels: []
milestone:
state: OPEN
stateReason:
createdAt: 2022-05-06T06:26:52Z
updatedAt: 2026-09-25T05:11:23Z
lastEditedAt: 2026-09-25T05:11:23Z
closedAt:
---

# add possibility to compare configuration of repos

Compare repository settings and show the deviations from a chosen reference. This helps to configure a family of projects identically and to spot differences that currently have to be hunted down on separate GitHub pages.

A provisional interface:

```bash
maintainer config compare octomation/maintainer kamilsk/retry
# issues: enabled → disabled
# merge settings: differ
```

The mandatory outcome is a comparison that changes nothing: which properties differ, what was taken as the reference, and which data could not be read. The set of compared settings has to be defined explicitly; lack of access is not the same as an absent setting.

Applying the difference was proposed as an **optional follow-up**:

```bash
maintainer config apply octomation/maintainer kamilsk/retry kamilsk/semaphore
```

It should rest on a reviewable plan ([[issue-19]]). Synchronizing template files is the different task [[issue-10]]; fetch does not change GitHub settings either.
