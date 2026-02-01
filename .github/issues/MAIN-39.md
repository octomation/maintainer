---
code: MAIN-39
id: I_kwDOE2M9Zc5JsYPz
databaseId: 1236370419
number: 39
url: https://github.com/octomation/maintainer/issues/39
title: "org: review and classify issues"
issueType: null
labels: []
assignees:
  - kamilsk
milestone: "[[MAIN-M2]]"
parent: null
issueFields: {}
state: CLOSED
stateReason: COMPLETED
duplicateOf: null
createdAt: 2022-05-15T19:11:31Z
updatedAt: 2026-10-21T10:22:07Z
lastEditedAt: 2026-10-19T07:22:40Z
closedAt: 2026-10-20T07:40:47Z
---

# org: review and classify issues

Bring the backlog to a state where it is clear what to do next and why. Terse titles, duplicates, and completed work mixed with live ideas all get in the way of assessing priorities.

The original actions: add meaningful prefixes, distribute the tasks across projects, close the unnecessary ones. Before changing any status, the code and the documents have to be checked: a closed issue may describe a removed feature, and an open one a partially implemented feature.

An example of the classification result:

```text
github: contribution: ... → a calendar defect, reproduction available
template: ...             → a maintenance improvement, needs a prototype
org: ...                  → an organizational outcome and release boundaries
```

Readiness means every task has an understandable motivation, an expected outcome, and its relations; duplicates explain where the work continues. External statuses and projects are changed by a separate action — editing Markdown locally does not update them.

<!-- 2026-10-20T07:40Z https://github.com/octomation/maintainer/issues/39#issuecomment-6285950064
Done: every task of the repository now belongs to one of three milestones, each with a concept that explains why its tasks are there.

- **GCC, v0.1** (milestone 1) — the v0.1 line: v0.1.0, its patches and everything before it. The line is about the contribution calendar; the upkeep helpers that v0.1.0 also ships (`go vanity build`, `makefile build`), the early imports and experiments and the dependency and CI upkeep of that time belong here because that is the release they shipped in or were dropped from.
- **Workspace, v0.2** (milestone 3) — `fetch` and `status`: v0.2.0 and the v0.2.x patches, `upstream` remotes for forks (#466) included.
- **Hub, v0.3** (milestone 2, formerly "Linear Graph") — a local web interface for repositories and tasks, synchronized with GitHub in v0.3; Linear, Jira and other trackers are the direction after it. One Homebrew Formula with the local service (#456) ships with it.

58 tasks changed milestone. Two tasks were closed as not planned because the milestone concepts superseded them: #52 (release slides, covered by the guides and release notes) and #53 (terminal dashboard, superseded by the Hub). Tasks whose relevance is disputed stay open in their milestones for a later decision: #11, #19, #48 and #197; #25 (comparing repository settings) moves to the Hub.
-->
