---
code:
id: I_kwDOE2M9Zc8AAAABTIUv1A
databaseId: 5578764244
number: 456
url: https://github.com/octomation/maintainer/issues/456
title: "release: homebrew: retire the formula in favour of the cask"
labels:
  - "type: improvement"
  - "impact: medium"
  - "effort: easy"
milestone:
state: OPEN
stateReason:
createdAt: 2026-09-25T05:49:59Z
updatedAt: 2026-09-25T05:49:59Z
lastEditedAt:
closedAt:
---

# release: homebrew: retire the formula in favour of the cask

Distribute maintainer through Homebrew as a cask only. GoReleaser deprecates `brews`, and a formula and a cask of the same name compete for one binary link, a conflict Homebrew cannot declare: a cask's `conflicts_with` accepts only casks, and a formula's only formulae.

The transition, in three steps:

1. Publish both. The formula carries a caveat and

   ```ruby
   deprecate! date: "2026-11-05", because: "is replaced by the cask", replacement_cask: "maintainer"
   ```

   The cask's `preflight_steps` warn when the formula owns the binary link and clear the quarantine attribute before the completions run the unsigned binary.
2. Retire the formula after the grace period: replace `deprecate!` with `disable!`, add `tap_migrations.json` with `{"maintainer": "maintainer"}` to octolab/homebrew-tap so that `brew upgrade` moves formula users to the cask, and delete `Formula/maintainer.rb`.
3. Remove `brews` from `.goreleaser.yml` and the accepted `brews` deprecation from `.github/scripts/goreleaser-check.mjs`; update the README, the documentation site and the workflows guide.

Expected: `brew upgrade` on a machine with the formula ends with the cask installed, linked and with completions; `make release-config-check` accepts no deprecation.

To confirm along the way: the first release that publishes both lands `Formula/maintainer.rb` and `Casks/maintainer.rb` in the tap, although GoReleaser warns that an artifact named `maintainer.rb` is already present; the cask installs on Linux.

Out of scope: signing and notarizing the macOS binary.

Related: [[issue-30]], [GoReleaser: brews deprecation](https://goreleaser.com/deprecations/#brews), [Homebrew: deprecating and disabling](https://docs.brew.sh/Deprecating-Disabling-and-Removing).
