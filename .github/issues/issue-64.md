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
updatedAt: 2026-09-25T06:45:26Z
lastEditedAt: 2026-09-25T05:13:09Z
closedAt:
---

# github: api: switch to graphql

Assess moving the GitHub integration to GraphQL where that makes fetching related data simpler and more reliable. The goal is to simplify the code — reducing maintenance complexity and the calendar's dependence on HTML changes — while preserving the user scenarios.

The reference check: fetch the calendar for a chosen period and compare the dates and counts against the current source. The `lookup`, `snapshot` and `suggest` commands must keep the meaning of their results regardless of how the data is loaded.

Data completeness, authorization rules, request limits and partial-failure handling all need checking; the advantages cannot be assumed from the name of the API. The official reference is the [GitHub GraphQL API](https://docs.github.com/en/graphql).

The [fetch specification](<../notes/Specs/GitHub fetcher, PoC implementation plan.md>) leaves GraphQL as a later experiment: the initial PoC must work over REST. This task must not turn GraphQL into a mandatory blocker for fetch.

<!-- 2026-09-25T06:40Z https://github.com/octomation/maintainer/issues/64#issuecomment-5828141481
Reference check of the calendar: GraphQL `contributionsCollection` against the year-scoped profile HTML, both read for `kamilsk` at 06:40Z on 2026-09-25.

| year | days | HTML total | GraphQL total | days that differ |
| --- | --- | --- | --- | --- |
| 2013 | 365 | 54 | 54 | none |
| 2019 | 365 | 1231 | 1229 | 2019-03-14: 2 vs 1; 2019-04-12: 6 vs 5 |
| 2021 | 365 | 1450 | 1449 | 2021-11-04: 8 vs 7 |
| 2026, up to 06:40:16Z | 268 | 5783 | 5783 | none |

- The dates agree in all four ranges. For the current year the HTML also lists the rest of the year as zero days, while GraphQL stops at `to`.
- The three differences are stable: the fixtures captured from the HTML earlier carry the same page values. For those days the GraphQL breakdown (commits, issues, pull requests, reviews) adds up to its own count and `restrictedContributionsCount` is 0, so the API does not itemise the extra contribution on the page. The query ran with the owner's gh CLI token; a contribution in a repository that this token cannot see is a likely explanation, not a confirmed one.
- `lookup`, `snapshot` and `suggest` read the same heat map, so their results can change only on such days.
-->
