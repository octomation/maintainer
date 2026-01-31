---
code: MAIN-139
id: I_kwDOE2M9Zc8AAAABYd-gxQ
databaseId: 5937012933
number: 491
url: https://github.com/octomation/maintainer/issues/491
title: "release: publishing: keep prereleases out of stable channels"
issueType: null
labels:
  - "type: improvement"
  - "scope: inventory"
  - "impact: high"
  - "effort: easy"
assignees: []
milestone: "[[MAIN-M3]]"
parent: null
issueFields: {}
state: OPEN
stateReason: null
duplicateOf: null
createdAt: 2026-10-20T06:25:19Z
updatedAt: 2026-10-20T06:26:10Z
lastEditedAt: 2026-10-20T06:26:10Z
closedAt: null
---

# release: publishing: keep prereleases out of stable channels

**Motivation:** v0.2.0 ships directly as a stable release, and the release guide allows no prerelease tag until this task is done. A later release may again be preceded by a candidate that the owner runs on the real collection before the final tag; a tag such as `v0.3.0-rc1` must not become the *Latest* release that `bin/install` resolves, and must not update the Homebrew packages that `brew upgrade` delivers. The tap preflight must not mistake a successful read for the ability to publish.

**Requirements**

- A tag with a prerelease suffix is published as a prerelease, never becomes *Latest*, and leaves every package in the tap untouched: the Formula and, while it is published, the Cask.
- The right to write to the tap is proved by the credential that publishes, not by a read: `cd` mints a token from the OctoLab Releaser App, scoped to the tap with contents write, in its "Mint the tap token" step before anything is published ([[MAIN-140]]), and `doctor` mints it the same way. A refused mint stops the release with a message that names the App installation and the tap; `permissions.push` of a personal token is no longer checked.
- `make release-config-check` still accepts exactly the `brews` deprecation and nothing else.

**Evidence:** findings 5 (assessed by reading the configuration and the GoReleaser source, consistent with `gh release view v0.1.0-rc11` reporting `"isPrerelease": false`) and 33 (reproduced with an unauthenticated read; a valid read-only token was not tested) of [the code review of 2026-09-26](../reviews/20260926T071417Z.md) — `.goreleaser.yml:93`, `.github/scripts/release.mjs:365`.

**PoC**

An artifact check in a disposable checkout, which builds but publishes nothing and creates no tag — GoReleaser takes the version from `GORELEASER_CURRENT_TAG`:

```bash
wt=$(mktemp -d) && git worktree add "$wt" HEAD || exit 1
( cd "$wt" && GORELEASER_CURRENT_TAG=v0.2.0-rc0 goreleaser release --snapshot --skip=publish --clean ); status=$?
git worktree remove --force "$wt"; exit $status
```

The publication promise itself is verified only by the first candidate after this task: before pushing it, record `gh release view --json tagName -q .tagName` of `releases/latest` and the checksums of `Formula/maintainer.rb` and, while it is published, `Casks/maintainer.rb` in the tap; after publication, `gh release view <candidate> --json isPrerelease` must be `true`, *Latest* must be unchanged, and those tap files must be byte-identical to the recorded ones.

**Out of scope:** which Homebrew packages are published ([[MAIN-125]]); signing or notarizing the binary; auditing other credentials.

Related: [[MAIN-138]].
