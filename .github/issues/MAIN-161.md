---
code: MAIN-161
id: null
databaseId: null
number: null
url: null
title: "go: vanity: remove the go vanity command"
issueType: Task
labels:
  - "scope: code"
  - "scope: docs"
  - "type: improvement"
assignees:
  - kamilsk
milestone: null
parent: null
issueFields: {}
state: OPEN
stateReason: null
duplicateOf: null
createdAt: null
updatedAt: null
lastEditedAt: null
closedAt: null
---

# go: vanity: remove the go vanity command

## Motivation

go.octolab.org moved from pull to push: every module publishes its own entry to
octolab/vanity with octomation/vanity-action, which reads the module from git and the
runner's `go`, with no maintainer. The last user of `maintainer go vanity build` is the
page rendering in octolab/vanity, and octolab/vanity#510 moves it into Python there. After
that the command serves nobody, and its code, docs and site section only add weight to
maintainer.

## PoC

```bash
grep -rln vanity internal/ docs/content skills/maintainer README.md
gh api repos/octolab/vanity/contents/.github/actions/setup/action.yml --jq .content | base64 -d | grep -n MAINTAINER
```

## Acceptance criteria

- [ ] octolab/vanity#510 is closed: octolab/vanity installs no maintainer.
- [ ] `maintainer go` and `maintainer go vanity` are gone with `internal/command/golang`
      and `internal/model/golang`; `go test ./...` passes.
- [ ] README, the site (`docs/content/vanity.md`, `_meta.js`, the «02 / GO» card on
      `index.mdx`, the link in `quickstart.md`), `docs/README.md` and the skill
      (`SKILL.md`, `references/cli.md`, `references/workflows.md`) no longer mention the
      command; the changelog of the next release names its removal.
- [ ] [[MAIN-4]], [[MAIN-21]] and [[MAIN-107]] are closed as not planned with a link here.

## Out of scope

- The `vanity` job of `cd.yml` and the go.octolab.org check of `doctor.yml` and
  `release.mjs`: they publish maintainer itself to the catalog and stay.
- Changelogs of past releases.
