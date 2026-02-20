---
code: MAIN-185
id: I_kwDOE2M9Zc8AAAABige2KQ
databaseId: 6610728489
number: 546
url: https://github.com/octomation/maintainer/issues/546
title: "ci: actions: pin external actions to a commit"
issueType: Bug
labels:
  - "scope: code"
  - "type: bug"
  - "severity: major"
  - "impact: high"
  - "effort: easy"
assignees:
  - kamilsk
milestone: null
parent: null
issueFields: {}
state: CLOSED
stateReason: COMPLETED
duplicateOf: null
createdAt: 2026-11-25T18:09:52Z
updatedAt: 2026-11-25T19:39:38Z
lastEditedAt: 2026-11-25T19:39:38Z
closedAt: 2026-11-25T18:44:44Z
---

# ci: actions: pin external actions to a commit

## Problem

Jobs that hold a release token or a write permission run third-party actions
by a tag that can be moved. The code that runs with these permissions must be
fixed to a reviewed commit.

## Expected behavior

- Every `uses:` of an action outside the owner's organization, in every
  workflow and composite action, `actions/*` included, is pinned to the full
  commit SHA of the tag in use, with the tag in a trailing comment:
  `uses: owner/action@<sha> # v1.2.3`.
- Actions of `octomation/*` stay pinned to a release tag: the tag is under the
  owner's control.
- Dependabot keeps updating the SHA pins and their comments.
- `.github/workflows/README.md` states the rule.

## Acceptance criteria

- [x] `grep -hn 'uses:' .github/workflows/*.yml .github/actions/*/action.yml`
      shows a 40-hex SHA for each action that is not `octomation/*` and not a
      local path.
- [x] Each SHA is the commit of the tag in its comment, checked through the
      GitHub API.
- [x] `actionlint` and `make tool-pins-check` pass.
- [x] `.github/workflows/README.md` states the rule.

## Out of scope

- Pinning `octomation/*` actions by SHA.
- No changelog entry: this changes the release pipeline, not the product.

Source: [the security review of 2026-11-25](../reports/20261125T152238Z.md), F5.
