---
code:
id: I_kwDOE2M9Zc5jGx7W
databaseId: 1662721750
number: 132
url: https://github.com/octomation/maintainer/issues/132
title: "linear: generate client"
labels:
  - "scope: code"
  - "type: feature"
  - "scope: deps"
  - "impact: high"
  - "effort: medium"
milestone: "[[milestone-2]]"
state: OPEN
stateReason:
createdAt: 2023-04-11T15:13:16Z
updatedAt: 2026-09-25T05:10:14Z
lastEditedAt: 2026-09-25T05:10:14Z
closedAt:
---

# linear: generate client

Prepare access to Linear issues as the enabler for a flat list and a tree of related issues. The user should see the structure of the work from the terminal instead of walking between cards by hand.

A provisional interface:

```bash
maintainer linear issues --flat
maintainer linear issues --tree
```

The minimal expected result: obtain the identifier, the title, the status and the parent/sub-task relations, and preserve them when switching between the two views. Access errors and partial loading must be visible; the tree must not lose issues that have no parent.

Generating a client is a means of enabling this scenario, not a value in itself without verified data retrieval. Original references: [genqlient](https://github.com/Khan/genqlient) and the [Linear schema](https://github.com/linear/linear/blob/master/packages/sdk/src/schema.graphql). The old `docs/INTRODUCTION.md` path of genqlient is no longer reachable, so the working project address is kept instead.

For ideas about a shared inbox and prioritization see [delayed automation](<../notes/Issues/delayed automation.md>); two-way synchronization is not automatically part of this first step.
