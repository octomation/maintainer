---
code: MAIN-139
id: null
databaseId: null
number: null
url: null
title: "release: publishing: keep prereleases out of stable channels"
labels:
  - "type: improvement"
  - "scope: inventory"
  - "impact: high"
  - "effort: easy"
milestone: "[[MAIN-M3]]"
state: OPEN
stateReason: null
createdAt: 2026-09-26T09:21:00Z
updatedAt: 2026-09-26T09:21:00Z
lastEditedAt: null
closedAt: null
---

# release: publishing: keep prereleases out of stable channels

**Motivation:** v0.2.0 is preceded by a release candidate that the owner runs on the real collection before the final tag. A `v0.2.0-rc1` tag must not become the *Latest* release that `bin/install` resolves, and must not update the Homebrew Formula and Cask that `brew upgrade` delivers. The tap preflight must not mistake a successful read for the ability to publish.

**Requirements**

- A tag with a prerelease suffix is published as a prerelease, never becomes *Latest*, and leaves the Formula and the Cask in the tap untouched.
- The preflight checks the token's effective capability as far as the API exposes it — `permissions.push` on the tap repository — and states its limit: it cannot prove that a fine-grained token may write; `release.test.mjs` covers a `permissions.push = false` response, and the release path includes one manual check with a token that can read but not write, when such a token is available.
- `make release-config-check` still accepts exactly the `brews` deprecation and nothing else.

**Evidence:** findings 5 (assessed by reading the configuration and the GoReleaser source, consistent with `gh release view v0.1.0-rc11` reporting `"isPrerelease": false`) and 33 (reproduced with an unauthenticated read; a valid read-only token was not tested) of [the code review of 2026-09-26](../reviews/20260926T071417Z.md) — `.goreleaser.yml:93`, `.github/scripts/release.mjs:365`.

**PoC**

An artifact check in a disposable checkout, which builds but publishes nothing and creates no tag — GoReleaser takes the version from `GORELEASER_CURRENT_TAG`:

```bash
wt=$(mktemp -d) && git worktree add "$wt" HEAD || exit 1
( cd "$wt" && GORELEASER_CURRENT_TAG=v0.2.0-rc0 goreleaser release --snapshot --skip=publish --clean ); status=$?
git worktree remove --force "$wt"; exit $status
```

The publication promise itself is verified only by the candidate: before pushing `v0.2.0-rc1`, record `gh release view --json tagName -q .tagName` of `releases/latest` and the checksums of `Formula/maintainer.rb` and `Casks/maintainer.rb` in the tap; after publication, `gh release view v0.2.0-rc1 --json isPrerelease` must be `true`, *Latest* must be unchanged, and both tap files must be byte-identical to the recorded ones.

**Out of scope:** retiring the Formula ([[MAIN-125]]); signing or notarizing the binary; auditing other credentials.

Related: [[MAIN-138]].
