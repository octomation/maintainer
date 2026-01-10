---
code:
id: I_kwDOE2M9Zc5Ll5Od
databaseId: 1268224925
number: 53
url: https://github.com/octomation/maintainer/issues/53
title: "tools: experimental: investigate cli dashboard for GitHub"
labels:
  - "scope: docs"
  - "scope: code"
  - "scope: test"
milestone:
state: OPEN
stateReason:
createdAt: 2022-06-11T10:02:54Z
updatedAt: 2023-01-06T13:48:47Z
lastEditedAt:
closedAt:
---

# tools: experimental: investigate cli dashboard for GitHub

Investigate a terminal dashboard for day-to-day work with GitHub, and where the contribution features belong in such an interface. The goal is to improve the developer experience by reducing switches between a browser and individual CLI calls.

The reference is [gh-dash](https://github.com/dlvhdr/gh-dash). A prototype of the user path: open the dashboard → choose the contributions section → look at the calendar → get a date suggestion.

The research must answer whether this scenario can be embedded into the `gh` environment — including how its extension mechanics work — how a user installs it, and what advantages it offers over the current commands. A minimal working example, or a justified decision not to proceed, is required; a list of libraries is not enough.

**Current state:** neither a dashboard nor an extension is shipped. The distribution mechanism is refined in [#97](issue-97.md). The local checkout-status table from the [status specification](<../notes/Specs/Repository status, PoC implementation plan.md>) serves a different scenario and should not be folded into this experiment automatically.
