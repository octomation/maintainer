---
code: MAIN-119
id: I_kwDOE2M9Zc58tL2F
databaseId: 2092219781
number: 197
url: https://github.com/octomation/maintainer/issues/197
title: "insights: deps: fetch data from deps.dev"
issueType: null
labels:
  - "scope: code"
  - "type: feature"
  - "impact: high"
  - "effort: medium"
assignees:
  - kamilsk
milestone: "[[MAIN-M1]]"
parent: null
issueFields: {}
state: OPEN
stateReason: null
duplicateOf: null
createdAt: 2024-01-20T19:51:39Z
updatedAt: 2026-11-20T15:42:30Z
lastEditedAt: 2026-11-20T15:42:30Z
closedAt: null
---

# insights: deps: fetch data from deps.dev

## Motivation

Judging a dependency means reading several sources by hand: which versions
exist, under which licence, what the package pulls in, and which advisories
are published for it. [deps.dev](https://deps.dev/) collects this for many
ecosystems and serves it through an API ([documentation](https://docs.deps.dev/)).
The idea was to bring that data into maintainer as an insight on a specific
package and version, starting from the [v3alpha API](https://docs.deps.dev/api/v3alpha/).

## PoC

```bash
$ maintainer insights deps go github.com/spf13/cobra@v1.10.2
# source:     deps.dev (API v3)
# licences:   Apache-2.0
# versions:   v1.10.2 (latest)
# deps:       3 direct, 5 transitive
# advisories: none published
$ maintainer insights deps go github.com/spf13/cobra@v1.10.2 --format json
```

## Acceptance criteria

- [ ] The ecosystem, the package and the version are given explicitly, and an
  unknown one is a clear error, not an empty report.
- [ ] The report shows the versions, the licences, the dependency tree and the
  advisories, and names deps.dev as their source.
- [ ] Absent data is reported as absent, never as "no problems".
- [ ] A JSON mode gives the same data for further processing.
- [ ] The API version is the one deps.dev supports at implementation time; the
  alpha contract is only the starting point.

## Out of scope

- Updating dependencies, or opening pull requests for them.
- Scanning a whole module or workspace: this is one package at a time.

<!-- 2026-11-20T21:00Z https://github.com/octomation/maintainer/issues/197#issuecomment-7043714324
Maturity: idea. Nothing in code: there is no insights group and no deps.dev client (`git grep -i deps.dev` finds only the issue mirror and reviews), and git history has no attempt.
GitHub's dependency graph and Dependabot already show versions, licences and advisories of a repository's dependencies.
Proposal for v0.5.0: not planned. Package risk analysis reads a package registry, not the contributor's repositories, issue trackers or GitHub work, so it is off the mission of the slimmed tool.
-->
