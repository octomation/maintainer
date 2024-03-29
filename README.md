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

...

## 🤼‍♂️ How to

...

## 🧩 Installation

### Homebrew

```bash
$ brew install octolab/tap/maintainer
```

### Binary

```bash
$ curl -sSfL https://raw.githubusercontent.com/octomation/maintainer/master/bin/install | sh
# or
$ wget -qO-  https://raw.githubusercontent.com/octomation/maintainer/master/bin/install | sh
```

> Don't forget about [security](https://www.idontplaydarts.com/2016/04/detecting-curl-pipe-bash-server-side/).

### Source

```bash
# use standard go tools
$ go get go.octolab.org/toolset/maintainer@latest
# or use egg tool
$ egg tools add go.octolab.org/toolset/maintainer@latest
```

> [egg][] is an `extended go get`.

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
[coverage.page]:    https://codeclimate.com/github/octomation/maintainer/test_coverage
[coverage.icon]:    https://api.codeclimate.com/v1/badges/6687c945bf44772d3131/test_coverage
[design.page]:      https://www.notion.so/octolab/maintainer-76d7f532a13244b5ac71708990f340ed
[docs.page]:        https://pkg.go.dev/go.octolab.org/toolset/maintainer
[docs.icon]:        https://img.shields.io/badge/docs-pkg.go.dev-blue
[mirror.page]:      https://bitbucket.org/kamilsk/maintainer
[mirror.icon]:      https://img.shields.io/badge/mirror-bitbucket-blue
[promo.page]:       https://github.com/octomation/maintainer
[quality.page]:     https://goreportcard.com/report/go.octolab.org/toolset/maintainer
[quality.icon]:     https://goreportcard.com/badge/go.octolab.org/toolset/maintainer
[template.page]:    https://github.com/octomation/go-tool
[template.icon]:    https://img.shields.io/badge/template-go--tool-blue

[egg]:              https://github.com/kamilsk/egg
