---
code:
id: I_kwDOE2M9Zc5Pr1gu
databaseId: 1336891438
number: 80
url: https://github.com/octomation/maintainer/issues/80
title: "github: labels: remove scope"
labels: []
milestone:
state: CLOSED
stateReason: COMPLETED
createdAt: 2022-08-12T08:09:14Z
updatedAt: 2023-03-31T15:31:20Z
lastEditedAt:
closedAt: 2023-03-31T15:31:20Z
---

# github: labels: remove scope

Remove GitHub label management from maintainer, since the settings app replaced that area. Two tools changing the same labels by different rules complicate maintenance and create conflicting expectations.

Expected result: the user-facing help no longer offers `github labels`, and the current examples do not point at removed operations. Historical tasks and material may be kept as long as their former context is stated explicitly.

**Current state:** the issue is closed; the [github](../../internal/command/github/root.go) group wires only the contribution commands and there is no label-management code. This explains the status of the older tasks about presets, dry-run and classification: [#8](issue-8.md), [#13](issue-13.md), [#14](issue-14.md), [#32](issue-32.md).

Removing labels does not cancel the general ideas of comparing configurations ([#25](issue-25.md)) and of reviewable transitions ([#19](issue-19.md)).
