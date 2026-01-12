---
title: Get started
description: Install maintainer, see your recent contribution pace, and choose a target.
---

# Get started

## Install

With Homebrew on macOS or Linux:

```sh
brew install --cask octolab/tap/maintainer
maintainer --help
```

The older Homebrew Formula is deprecated on 2026-11-05. If you have it, switch with `brew uninstall --formula maintainer && brew install --cask octolab/tap/maintainer`, and keep only one installation.

Without Homebrew, the install script puts the latest release into `./bin`; `-b` chooses another directory and a tag pins the version:

```sh
curl -sSfL https://raw.githubusercontent.com/octomation/maintainer/main/bin/install | sh -s -- -b ~/.local/bin v0.1.0
```

`wget -qO-` works in place of `curl -sSfL`. Archives for macOS and Linux on amd64 and arm64 are also on [GitHub Releases](https://github.com/octomation/maintainer/releases). With Go 1.27 or newer, no checkout is needed:

```sh
go install go.octolab.org/toolset/maintainer@latest
```

`go install` puts the executable in `GOBIN`, or `$(go env GOPATH)/bin` if `GOBIN` is unset; that directory must be on your `PATH`. From a checkout, `go install .` builds the current source.

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
maintainer github contribution lookup
maintainer github contribution suggest --target 25 2026-01-12/-1
```

Without an argument, `lookup` shows the week of the latest commit in the current repository and the week before it, or the current week outside a repository. A real excerpt from `suggest`, anchored at a past date:

```text
 Day / Week    #02    #03    Date
 Monday         25    17*   Jan 12

Suggestion is 2026-01-12T08:09:16+03:00, 17 → 25
```

`suggest` looks for the first moment from the anchor up to now, between 05:00 and 19:00 UTC, on a day below its target, and marks that day with `*`: this example shows 17 contributions toward a minimum target of 25. Your counts and suggested time will differ, but the time is never later than now; if no such moment exists, `suggest` fails instead. The time is a suggestion; maintainer does **not** create a commit.

Next: [read the contribution guide](/contributions/) to choose date ranges, inspect your activity distribution, and compare snapshots when you need a record. The [Go vanity URL](/vanity/) and [Makefile](/makefiles/) guides cover the smaller project helpers.
