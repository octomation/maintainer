---
code: MAIN-140
id: null
databaseId: null
number: null
url: null
title: "release: publishing: write the tap through the OctoLab Releaser App"
labels:
  - "type: improvement"
  - "scope: inventory"
milestone: null
state: OPEN
stateReason: null
createdAt: null
updatedAt: null
lastEditedAt: null
closedAt: null
---

# release: publishing: write the tap through the OctoLab Releaser App

**Motivation:** `cd` pushes the Formula and the Cask to `octolab/homebrew-tap` with `HOMEBREW_TAP_TOKEN`, a fine-grained personal access token owned by `octolab` (`.github/workflows/cd.yml:40,76`, `.goreleaser.yml:26,56`). Commits made with it are not verified, and it is the last user of that token: octopot/indexit already mints a short-lived token from the GitHub App "OctoLab Releaser" and gets verified tap commits (octopot/indexit#108). The same App now also publishes agent skills to `octolab/skills`. Its credentials are stored as `OCTOLAB_RELEASER_CLIENT_ID` and `OCTOLAB_RELEASER_KEY`.

**Requirements**

- `cd` writes the Formula and the Cask with a token minted from the App, scoped to the tap repository with contents write access; no step reads `HOMEBREW_TAP_TOKEN`.
- Tap commits of the next release are `verified: true` and authored by the App's bot.
- `doctor` and the release preflight check the App secrets and the App's access to the tap instead of the token, and explain how to fix a missing secret or installation.
- `.github/workflows/README.md` describes the App secrets; the token is gone from it.
- After a release proves the new path, the owner deletes the `HOMEBREW_TAP_TOKEN` secret and revokes the token at github.com/settings/personal-access-tokens; nothing else in the organization may still use it.

**PoC**

What octopot/indexit does, from its `cd.yml`:

```yaml
- name: Mint the tap token
  id: tap
  uses: actions/create-github-app-token@v3.2.0
  with:
    client-id: ${{ secrets.OCTOLAB_RELEASER_CLIENT_ID }}
    private-key: ${{ secrets.OCTOLAB_RELEASER_KEY }}
    owner: octolab
    repositories: homebrew-tap
    permission-contents: write
```

with `HOMEBREW_TAP_TOKEN: ${{ steps.tap.outputs.token }}` passed to GoReleaser and `commit_author.use_github_app_token: true` on the Cask. Whether the Formula (`brews`) gets a verified commit the same way is unchecked: run `goreleaser release --snapshot --skip=publish --clean` to validate the configuration, and confirm verification on the first real release with `gh api repos/octolab/homebrew-tap/commits/<sha> --jq .commit.verification`.

Before revoking the token, search the organization for remaining users: `gh search code HOMEBREW_TAP_TOKEN --owner octolab --owner octomation --owner octopot`.

**Out of scope:** keeping prereleases out of the tap ([[MAIN-139]]); retiring the Formula ([[MAIN-125]]); publishing agent skills to `octolab/skills`.
