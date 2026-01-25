---
code: MAIN-80
id: I_kwDOE2M9Zc5Pr1gu
databaseId: 1336891438
number: 80
url: https://github.com/octomation/maintainer/issues/80
title: "github: labels: remove scope"
issueType: null
labels: []
assignees:
  - kamilsk
milestone: null
parent: null
issueFields: {}
state: CLOSED
stateReason: COMPLETED
duplicateOf: null
createdAt: 2022-08-12T08:09:14Z
updatedAt: 2026-09-25T05:13:53Z
lastEditedAt: 2026-09-25T05:13:53Z
closedAt: 2023-03-31T15:31:20Z
---

# github: labels: remove scope

Remove GitHub label management from maintainer, since the settings app replaced that area. Two tools changing the same labels by different rules complicate maintenance and create conflicting expectations.

Expected result: the user-facing help no longer offers `github labels`, and the current examples do not point at removed operations. Historical tasks and material may be kept as long as their former context is stated explicitly.

Removing labels does not cancel the general ideas of comparing configurations ([[MAIN-25]]) and of reviewable transitions ([[MAIN-19]]).

Related: the older tasks about presets, dry-run and classification [[MAIN-8]], [[MAIN-13]], [[MAIN-14]], [[MAIN-32]].
