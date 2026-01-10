---
code:
id: I_kwDOE2M9Zc5NDnsu
databaseId: 1292794670
number: 64
url: https://github.com/octomation/maintainer/issues/64
title: "github: api: switch to graphql"
labels:
  - "scope: code"
milestone:
state: OPEN
stateReason:
createdAt: 2022-07-04T08:10:52Z
updatedAt: 2026-09-25T05:13:09Z
lastEditedAt: 2026-09-25T05:13:09Z
closedAt:
---

# github: api: switch to graphql

Assess moving the GitHub integration to GraphQL where that makes fetching related data simpler and more reliable. The goal is to simplify the code — reducing maintenance complexity and the calendar's dependence on HTML changes — while preserving the user scenarios.

The reference check: fetch the calendar for a chosen period and compare the dates and counts against the current source. The `lookup`, `snapshot` and `suggest` commands must keep the meaning of their results regardless of how the data is loaded.

Data completeness, authorization rules, request limits and partial-failure handling all need checking; the advantages cannot be assumed from the name of the API. The official reference is the [GitHub GraphQL API](https://docs.github.com/en/graphql).

The [fetch specification](<../notes/Specs/GitHub fetcher, PoC implementation plan.md>) leaves GraphQL as a later experiment: the initial PoC must work over REST. This task must not turn GraphQL into a mandatory blocker for fetch.
