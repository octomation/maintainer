> # 👨‍🔧 I'm a maintainer
>
> Contribution assistant for open source projects.

[![Build][build.icon]][build.page]
[![Documentation][docs.icon]][docs.page]
[![Quality][quality.icon]][quality.page]
[![Template][template.icon]][template.page]
[![Coverage][coverage.icon]][coverage.page]
[![Mirror][mirror.icon]][mirror.page]

## 💡 Idea

```bash
$ maintainer go vanity build
```

## 🏆 Motivation

Inspect GitHub contribution calendars and automate recurring repository maintenance.

## 🤼‍♂️ How to

Start with `maintainer --help` and the [documentation][docs.page].
See [development tools](tools/README.md) and [GitHub workflows](.github/workflows/README.md) for repository checks and releases.

## 🧩 Installation

### Homebrew

```bash
$ brew install --cask octolab/tap/maintainer
```

The cask is available for macOS and Linux. The formula
(`brew install --formula octolab/tap/maintainer`) keeps existing installations updated
and is deprecated on 2026-11-05. Keep only one of them; to switch, run
`brew uninstall --formula maintainer && brew install --cask maintainer`.

### Binary

```bash
$ curl -sSfL https://raw.githubusercontent.com/octomation/maintainer/main/bin/install | sh
# or
$ wget -qO-  https://raw.githubusercontent.com/octomation/maintainer/main/bin/install | sh
```

The script installs the latest release into `./bin`. Pass `-b` to choose another directory and a tag to pin the version:

```bash
$ curl -sSfL https://raw.githubusercontent.com/octomation/maintainer/main/bin/install | sh -s -- -b ~/.local/bin v0.1.0
```

> Don't forget about [security](https://www.idontplaydarts.com/2016/04/detecting-curl-pipe-bash-server-side/).

### Source

Install the latest release with Go 1.27 or newer, no checkout needed:

```bash
$ go install go.octolab.org/toolset/maintainer@latest
```

The executable is installed into `GOBIN`, or `$(go env GOPATH)/bin` when `GOBIN` is unset. Add that directory to `PATH`.

### Shell completions

```bash
$ maintainer completion > /path/to/completions/...
# or
$ source <(maintainer completion)
```

<p align="right">made with ❤️ for everyone</p>

[awesome.icon]:     https://awesome.re/mentioned-badge.svg
[build.page]:       https://github.com/octomation/maintainer/actions/workflows/ci.yml
[build.icon]:       https://github.com/octomation/maintainer/actions/workflows/ci.yml/badge.svg
[coverage.page]:    https://app.codecov.io/gh/octomation/maintainer
[coverage.icon]:    https://codecov.io/gh/octomation/maintainer/branch/main/graph/badge.svg
[design.page]:      https://www.notion.so/octolab/maintainer-76d7f532a13244b5ac71708990f340ed
[docs.page]:        https://maintainer.octolab.org/
[docs.icon]:        https://img.shields.io/badge/docs-GitHub%20Pages-blue
[mirror.page]:      https://bitbucket.org/kamilsk/maintainer
[mirror.icon]:      https://img.shields.io/badge/mirror-bitbucket-blue
[promo.page]:       https://github.com/octomation/maintainer
[quality.page]:     https://goreportcard.com/report/go.octolab.org/toolset/maintainer
[quality.icon]:     https://goreportcard.com/badge/go.octolab.org/toolset/maintainer
[template.page]:    https://github.com/octomation/go-tool
[template.icon]:    https://img.shields.io/badge/template-go--tool-blue

[egg]:              https://github.com/kamilsk/egg
