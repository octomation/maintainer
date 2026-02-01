---
code: MAIN-125
id: I_kwDOE2M9Zc8AAAABTIUv1A
databaseId: 5578764244
number: 456
url: https://github.com/octomation/maintainer/issues/456
title: "release: homebrew: ship one formula with the service and retire the cask"
issueType: null
labels:
  - "type: improvement"
  - "impact: medium"
  - "effort: easy"
assignees:
  - kamilsk
milestone: "[[MAIN-M3]]"
parent: null
issueFields: {}
state: OPEN
stateReason: null
duplicateOf: null
createdAt: 2026-09-25T05:49:59Z
updatedAt: 2026-10-21T10:22:30Z
lastEditedAt: 2026-10-21T05:15:39Z
closedAt: null
---

# release: homebrew: ship one formula with the service and retire the cask

Distribute maintainer through Homebrew as one Formula that carries the CLI, its shell completions and the `brew services` definition of `maintainer service run` ([[MAIN-151]]), as indexit does. The CLI and the service are one binary and upgrade together; a separate formula for the service alone is not an option. `brew services` manages formulae only: a cask has no launchd stanza, and GoReleaser's `homebrew_casks.service` is a macOS Services item, not a launchd job. The only GoReleaser generator of a Formula is `brews`: deprecated, but supported until GoReleaser v3. The Cask, which v0.2.x publishes next to the Formula ([[MAIN-152]]), is retired.

The transition, in two steps:

1. v0.3.0 publishes the Formula through `brews` with `install`, `service`, `caveats` and `test`, still without the `deprecate!` that [[MAIN-152]] lifts. `homebrew_casks` leaves `.goreleaser.yml`, so the Cask stays at its last v0.2.x. `make release-config-check` keeps accepting exactly the `brews` deprecation.
2. Once that release is out, a separate reviewed change in octolab/homebrew-tap deletes `Casks/maintainer.rb`; GoReleaser does not delete the old Cask. The Cask has no users besides the maintainer, who moves by hand, as the service guide describes:

   ```sh
   brew uninstall --cask maintainer && brew install --formula octolab/tap/maintainer
   ```

Formula users get the new Formula with `brew upgrade`; nothing else changes for them.

Expected: `brew services` runs the local service from the same Formula as the CLI; after the move, the maintainer's machine has the Formula installed and linked, with completions, and no Cask; `make release-config-check` accepts the `brews` deprecation and nothing else.

## Acceptance criteria

- [ ] `make go-dist-check` generates `Formula/maintainer.rb` with `service do` running `maintainer service run`, caveats that name `http://maintainer.localhost:17841/` and `brew services start octolab/tap/maintainer`, and no `deprecate!`; no Cask is generated.
- [ ] `make release-config-check` passes with the `brews` deprecation and fails on any other deprecation or an invalid config.
- [ ] After the release, on a Mac with a v0.2.x Formula, `brew upgrade` installs the new Formula, and after `brew services start octolab/tap/maintainer` the service answers at `http://maintainer.localhost:17841/`.
- [ ] On a Mac with a v0.2.x Cask, the manual move ends with the Formula linked, with completions, and no Cask; `Casks/maintainer.rb` is gone from the tap.
- [ ] README, Get started, the service guide, AGENTS.md and the workflows guide install and upgrade through the Formula, and name the Cask only for the move; the published v0.2.0 note stays as it is.

Out of scope: signing and notarizing the macOS binary; a Linux service; keeping prereleases out of the tap ([[MAIN-139]]).

Related: [[MAIN-30]], [[MAIN-151]], [[MAIN-152]], [GoReleaser: brews deprecation](https://goreleaser.com/deprecations/#brews), [Homebrew: deprecating and disabling](https://docs.brew.sh/Deprecating-Disabling-and-Removing).
