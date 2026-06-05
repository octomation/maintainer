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

## 🛰️ `maintainer fetch`

`maintainer fetch` discovers GitHub repositories across several owners and
reconciles a local checkout tree, in the spirit of `terraform plan` /
`terraform apply`. A local state file (keyed by the stable numeric repo `id`)
remembers what was materialised, so a rename or transfer on GitHub is detected
as a **move**, not as a delete-and-reclone.

It is **plan-only by default** and **safe-by-default on disk**: `--apply`
performs only non-destructive actions (clone, fetch refs, move, update remote,
adopt). A repository that disappears from GitHub is reported as an `orphan`
(404-confirmed) and the local clone is **retained, never deleted**.

```bash
# scaffold a documented config, then check it
$ maintainer fetch config init           # writes ./fetch.toml
$ maintainer fetch config validate

# render a plan (no disk writes), then apply non-destructive actions
$ maintainer fetch                       # plan only
$ maintainer fetch --apply               # clone / fetch / move / update-remote / adopt

# machine-readable plan for a wrapping tool (lists every action incl. fetches)
$ maintainer fetch --format=json | jq .

# scope a run; inspect or tidy state
$ maintainer fetch --profile=primary --owner=acme
$ maintainer fetch state show            # dump state.json
$ maintainer fetch state prune           # forget records whose path is gone

# single-run mode without a config file (a token is required)
$ GITHUB_TOKEN=ghp_… maintainer fetch --owner=acme --apply
```

Configuration lives in `fetch.{toml,yaml}` (`defaults`, `filters`,
`[profiles.<name>]`, `[[owners]]`, `[[repos]]`); see the template written by
`fetch config init`. Per-profile tokens make a bot account's private repos
reachable (`clone_url = "https"` + its own `token_env`). The state file
defaults to `$XDG_STATE_HOME/maintainer/fetch/state.json` (`0600`).

Exit codes: `0` clean (incl. "no drift"), `1` transport/Git/state error,
`2` user input error (bad config/flags, missing token), `3` apply finished with
at least one per-repo failure (the summary lists which).

Full reference: [`docs/fetch.md`](docs/fetch.md).

> The PoC ships REST discovery only; a GraphQL discoverer is a deferred
> experiment. See [the PoC plan](.github/notes/) for the full design.

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
