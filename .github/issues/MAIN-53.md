---
code: MAIN-53
id: I_kwDOE2M9Zc5Ll5Od
databaseId: 1268224925
number: 53
url: https://github.com/octomation/maintainer/issues/53
title: "tools: experimental: investigate cli dashboard for GitHub"
issueType: null
labels:
  - "scope: docs"
  - "scope: code"
  - "scope: test"
assignees:
  - kamilsk
milestone: "[[MAIN-M2]]"
parent: null
issueFields: {}
state: CLOSED
stateReason: NOT_PLANNED
duplicateOf: null
createdAt: 2022-06-11T10:02:54Z
updatedAt: 2026-10-20T07:40:42Z
lastEditedAt: 2026-10-19T07:22:48Z
closedAt: 2026-10-20T07:40:42Z
---

# tools: experimental: investigate cli dashboard for GitHub

Investigate a terminal dashboard for day-to-day work with GitHub, and where the contribution features belong in such an interface. The goal is to improve the developer experience by reducing switches between a browser and individual CLI calls.

The reference is [gh-dash](https://github.com/dlvhdr/gh-dash). A prototype of the user path: open the dashboard → choose the contributions section → look at the calendar → get a date suggestion.

The research must answer whether this scenario can be embedded into the `gh` environment — including how its extension mechanics work — how a user installs it, and what advantages it offers over the current commands. A minimal working example, or a justified decision not to proceed, is required; a list of libraries is not enough.

The local checkout-status table of [`maintainer status`](https://maintainer.octolab.org/status/) serves a different scenario and should not be folded into this experiment automatically.

Related: [[MAIN-91]].

<!-- 2026-10-20T07:40Z https://github.com/octomation/maintainer/issues/53#issuecomment-6285948488
Closing as not planned following the Hub milestone's choice of a local web interface for repositories and tasks. The planned interface runs locally through `brew services`, with GitHub synchronization in v0.3 and other trackers later; #456 covers distributing the CLI and the service together. This replaces the proposed terminal dashboard as the product direction, without implying that a contribution-calendar page is already implemented or committed. GitHub CLI extension distribution remains a separate experiment in #97, and `maintainer status` continues to address local checkout inspection.
-->
