---
title: Get started
description: Install maintainer, see your recent contribution pace, and choose a target.
---

# Get started

## Install

Published builds are on [GitHub Releases](https://github.com/octomation/maintainer/releases). Choose the archive matching your OS and architecture; v0.1.0 is prepared for macOS and Linux on amd64 and arm64. The v0.1.0 release also adds a Homebrew Cask:

```sh
brew install --cask octolab/tap/maintainer
maintainer --help
```

If you already installed the older Homebrew Formula, switch with `brew uninstall --formula maintainer && brew install --cask octolab/tap/maintainer`. Keep only one installation. To build the current source from a checkout instead:

```sh
git clone https://github.com/octomation/maintainer.git
cd maintainer
go install .
maintainer --help
```

Use Go 1.27 or newer. `go install .` puts the executable in `GOBIN`, or `$(go env GOPATH)/bin` if `GOBIN` is unset; that directory must be on your `PATH`.

## Give it a GitHub token

Contribution commands query GitHub's GraphQL API for **the account owning the token**. A token is required even when that profile is public. Use either form:

```sh
# Option 1: reuse the token in this shell
export GITHUB_TOKEN=your_token

# Option 2: pass it for one command
maintainer github --token=your_token contribution lookup now/-3
```

The environment variable applies to later commands in the same shell. `--token` applies to one command and may remain in shell history. Private contribution details depend on your GitHub permissions and settings.

## Plan your next contribution

```sh
maintainer github contribution lookup now/-3
maintainer github contribution suggest --target 25 now/-3
```

An illustrative excerpt from `suggest`:

```text
 Day / Week   #36   #37   #38   #39    Date
 Friday       25    25    25    5*    Sep 25

Suggestion is 2026-09-25T12:46:48+03:00, 5 → 25
```

`lookup` puts the last three weeks beside the current one so you can see your recent pace. `suggest` marks a candidate day with `*`: this example shows five contributions toward a minimum target of 25. Your counts and suggested time will differ. The time is a suggestion, not a scheduled action; maintainer does **not** create a commit.

Next: [read the contribution guide](/contributions/) to choose date ranges, inspect your activity distribution, and compare snapshots when you need a record. The [Go vanity URL](/vanity/) and [Makefile](/makefiles/) guides cover the smaller project helpers.
