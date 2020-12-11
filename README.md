> # 👨‍🔧 maintainer
>
> Toolset for Open Source contribution.

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

A full description of the idea is available [here][design.page].

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
$ maintainer completion bash|fish|powershell|zsh > /path/to/completions/...
```

> See `kubectl` [documentation](https://kubernetes.io/docs/tasks/tools/install-kubectl/#enabling-shell-autocompletion).

<p align="right">made with ❤️ for everyone</p>

[build.page]:       https://travis-ci.com/octomation/maintainer
[build.icon]:       https://travis-ci.com/octomation/maintainer.svg?branch=master
[coverage.page]:    https://codeclimate.com/github/octomation/maintainer/test_coverage
[coverage.icon]:    https://api.codeclimate.com/v1/badges/6687c945bf44772d3131/test_coverage
[design.page]:      https://www.notion.so/octolab/maintainer-76d7f532a13244b5ac71708990f340ed?r=0b753cbf767346f5a6fd51194829a2f3
[docs.page]:        https://pkg.go.dev/go.octolab.org/toolset/maintainer
[docs.icon]:        https://img.shields.io/badge/docs-pkg.go.dev-blue
[promo.page]:       https://github.com/octomation/maintainer
[quality.page]:     https://goreportcard.com/report/go.octolab.org/toolset/maintainer
[quality.icon]:     https://goreportcard.com/badge/go.octolab.org/toolset/maintainer
[template.page]:    https://github.com/octomation/go-tool
[template.icon]:    https://img.shields.io/badge/template-go--tool-blue
[mirror.page]:      https://bitbucket.org/kamilsk/maintainer
[mirror.icon]:      https://img.shields.io/badge/mirror-bitbucket-blue

[egg]:              https://github.com/kamilsk/egg
