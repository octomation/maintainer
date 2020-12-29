> # 🧩 Tool
>
> Template for typical Go tool.

[![Build][build.icon]][build.page]
[![Documentation][docs.icon]][docs.page]
[![Quality][quality.icon]][quality.page]
[![Template][template.icon]][template.page]
[![Coverage][coverage.icon]][coverage.page]
[![Mirror][mirror.icon]][mirror.page]

## 💡 Idea

```bash
$ tool do action
```

A full description of the idea is available [here][design.page].

## 🏆 Motivation

...

## 🤼‍♂️ How to

...

## 🧩 Installation

### Homebrew

```bash
$ brew install :owner/tap/:binary
```

### Binary

```bash
$ curl -sSfL https://raw.githubusercontent.com/:owner/:repository/master/bin/install | sh
# or
$ wget -qO-  https://raw.githubusercontent.com/:owner/:repository/master/bin/install | sh
```

> Don't forget about [security](https://www.idontplaydarts.com/2016/04/detecting-curl-pipe-bash-server-side/).

### Source

```bash
# use standard go tools
$ go get github.com/:owner/:repository@:version
# or use egg tool
$ egg tools add github.com/:owner/:repository@:version
```

> [egg][] is an `extended go get`.

### Bash and Zsh completions

```bash
$ :binary completion bash > /path/to/bash_completion.d/:binary.sh
$ :binary completion zsh  > /path/to/zsh-completions/_:binary.zsh
# or autodetect
$ source <(:binary completion)
```

> See `kubectl` [documentation](https://kubernetes.io/docs/tasks/tools/install-kubectl/#enabling-shell-autocompletion).

## 🤲 Outcomes

...

---

made with ❤️ for everyone

[build.page]:       https://travis-ci.com/:owner/:repository
[build.icon]:       https://travis-ci.com/:owner/:repository.svg?branch=master
[coverage.page]:    https://codeclimate.com/github/:owner/:repository/test_coverage
[coverage.icon]:    https://api.codeclimate.com/v1/badges/c570179a9335c747e77c/test_coverage
[design.page]:      https://www.notion.so/33715348cc114ea79dd350a25d16e0b0?r=0b753cbf767346f5a6fd51194829a2f3
[docs.page]:        https://pkg.go.dev/:module/:version
[docs.icon]:        https://img.shields.io/badge/docs-pkg.go.dev-blue
[promo.page]:       https://github.com/:owner/:repository
[quality.page]:     https://goreportcard.com/report/:module
[quality.icon]:     https://goreportcard.com/badge/go.octolab.org
[template.page]:    https://github.com/octomation/go-tool
[template.icon]:    https://img.shields.io/badge/template-go--tool-blue
[mirror.page]:      https://bitbucket.org/kamilsk/go-tool
[mirror.icon]:      https://img.shields.io/badge/mirror-bitbucket-blue

[egg]:              https://github.com/kamilsk/egg
