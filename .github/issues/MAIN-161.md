---
code: MAIN-161
id: I_kwDOE2M9Zc8AAAABhOjt0A
databaseId: 6524825040
number: 536
url: https://github.com/octomation/maintainer/issues/536
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
createdAt: 2026-11-21T05:16:38Z
updatedAt: 2026-11-21T05:17:12Z
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
gh search code --repo octolab/vanity maintainer --json path --jq .[].path   # no workflow or script installs it
```

## Acceptance criteria

- [ ] octolab/vanity#510 is closed: octolab/vanity installs no maintainer.
- [x] `maintainer go` and `maintainer go vanity` are gone with `internal/command/golang`
      and `internal/model/golang`; `go test ./...` passes.
- [x] README, the site (`docs/content/vanity.md`, `_meta.js`, the «02 / GO» card on
      `index.mdx`, the link in `quickstart.md`), `docs/README.md` and the skill
      (`SKILL.md`, `references/cli.md`, `references/workflows.md`) no longer mention the
      command; the changelog of the next release names its removal.
- [ ] [[MAIN-4]], [[MAIN-21]] and [[MAIN-107]] are closed as not planned with a link here.

## Out of scope

- The `vanity` job of `cd.yml` and the go.octolab.org check of `doctor.yml` and
  `release.mjs`: they publish maintainer itself to the catalog and stay.
- Changelogs of past releases.

<!-- 2026-11-20T19:00Z https://github.com/octomation/maintainer/issues/536#issuecomment-7057806937
Prepared in the working tree, to ship with the next release; octolab/vanity#510 is done in its working tree and waits for the push.

- `internal/command/golang` and `internal/model/golang` are gone, `maintainer go` with them; `go.mod` is unchanged (afero, yaml.v2 and go.octolab.org/strings are still used by `internal/config`); `go build`, `go vet` and `go test ./...` pass.
- README shows `maintainer status` in place of `go vanity build`; the site has no `vanity` page, menu item, home card or quick start link, and `next build` passes; `docs/README.md` no longer lists the page.
- The skill: the row in `SKILL.md` and the example in `references/workflows.md` are gone, `references/cli.md` is regenerated with `go generate ./skills`.
- Left: the changelog of the next release (written when the release is prepared), and closing MAIN-4, MAIN-21 and MAIN-107 as not planned when this lands.
-->
