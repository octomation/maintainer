---
code: MAIN-M1
id: MI_kwDOE2M9Zc4Ae1gL
databaseId: 8083467
number: 1
url: https://github.com/octomation/maintainer/milestone/1
title: "GCC, v0.1"
state: OPEN
openIssueCount: 24
closedIssueCount: 90
dueOn: null
createdAt: 2022-06-15T10:12:33Z
updatedAt: 2026-10-20T07:40:37Z
closedAt: null
---

# GCC, v0.1

## GitHub Contributions Calendar

Add commands to work with [GCC](https://docs.github.com/en/account-and-profile/setting-up-and-managing-your-github-profile/managing-contribution-graphs-on-your-profile/viewing-contributions-on-your-profile#contributions-calendar).

The milestone carries the v0.1 line: v0.1.0, its patches, and everything the project did before it. The line is about the contribution calendar; the rest of its tasks — the Go and Makefile upkeep helpers that v0.1.0 also ships, the early imports and experiments, the dependency and CI upkeep of that time — belong here because that is the release they shipped in or were dropped from.

## Issues

- [x] [[MAIN-1]] — import go code from vanity project
- [x] [[MAIN-2]] — implement doctoc
- [x] [[MAIN-3]] — import go code from makefiles project
- [ ] [[MAIN-4]] — how to warmup godoc and vanity url
- [ ] [[MAIN-5]] — integrate or implement embedmd
- [x] [[MAIN-6]] — issue with modules without go files on the root
- [x] [[MAIN-7]] — remove double new lines
- [x] [[MAIN-8]] — big picture: labels
- [x] [[MAIN-9]] — research: automate update golangci-lint config
- [ ] [[MAIN-10]] — template: automate sync with upstream
- [ ] [[MAIN-11]] — integrate markdown parser
- [x] [[MAIN-12]] — fetch current state of labels of all repositories
- [x] [[MAIN-13]] — label presets
- [x] [[MAIN-14]] — add dry-run option for github labels
- [x] [[MAIN-15]] — extend octolab preset by other repositories
- [x] [[MAIN-16]] — org: prepare release v0.2.0
- [x] [[MAIN-17]] — extend gh tool
- [x] [[MAIN-18]] — update github client
- [x] [[MAIN-20]] — contribution chart
- [ ] [[MAIN-21]] — improve vanity support
- [x] [[MAIN-22]] — update github client
- [x] [[MAIN-23]] — sync with latest template
- [x] [[MAIN-24]] — use contributions chart for git at suggestion
- [x] [[MAIN-26]] — implement github contribution lookup command
- [x] [[MAIN-27]] — implement github contribution snapshot command
- [x] [[MAIN-28]] — implement github contribution diff command
- [x] [[MAIN-29]] — implement github contribution histogram command
- [x] [[MAIN-30]] — org: prepare release v0.1.0
- [x] [[MAIN-32]] — command: labels classify issue
- [x] [[MAIN-33]] — provide github access token by command flag
- [x] [[MAIN-34]] — disable hub subcommands
- [x] [[MAIN-35]] — define tool config
- [x] [[MAIN-36]] — command: github contribution lookup doesn't work well with now()
- [x] [[MAIN-37]] — deps: up go-github to v44.1.0
- [x] [[MAIN-38]] — github: contribution: lookup shows incorrect scope for 1 week with now ts
- [x] [[MAIN-40]] — github: contribution: extend heat map by .From(), .To(), and .Range()
- [x] [[MAIN-41]] — github: contribution: lookup forward/backward
- [x] [[MAIN-42]] — deps: up go-github to v45.0.0
- [x] [[MAIN-43]] — github: contribution: expand heat map by merging with neighbors
- [ ] [[MAIN-44]] — github: contribution: refactor command and view
- [ ] [[MAIN-45]] — github: contribution: extend support lookup for a month
- [x] [[MAIN-46]] — github: contribution: show datetime on the right side of lookup
- [x] [[MAIN-47]] — org: define milestones
- [x] [[MAIN-49]] — deps: up go-github to v45.1.0
- [ ] [[MAIN-50]] — command: separate command structure and command execution
- [ ] [[MAIN-51]] — di: define service provider to inject it into command for lazy service initialization
- [x] [[MAIN-52]] — tools: experimental: integrate slides and describe v0.1.x using it
- [x] [[MAIN-54]] — ci/cd: invalid workflow file: .github/workflows/cd.yml#L38
- [x] [[MAIN-55]] — ci/cd: rename coverage to report
- [x] [[MAIN-56]] — github: contribution: support --short for suggest
- [x] [[MAIN-57]] — github: contribution: refactor logic of suggest command
- [x] [[MAIN-58]] — github: contribution: add --auto flag to suggest command
- [x] [[MAIN-59]] — github: contribution: prevent caching of github contribution calendar
- [x] [[MAIN-60]] — github: contribution: replace --weeks by /weeks argument format for suggest
- [x] [[MAIN-61]] — text: replace -> by →
- [x] [[MAIN-62]] — deps: up github.com/google/go-github to v45.2.0
- [x] [[MAIN-63]] — github: contribution: highlight suggested day
- [x] [[MAIN-64]] — github: api: switch to graphql
- [x] [[MAIN-65]] — github: contribution: lookup has problem with timezone
- [x] [[MAIN-66]] — github: contribution: lookup doesn't show last week on Sunday
- [x] [[MAIN-67]] — github: contribution: suggest doesn't work on last week with Sunday
- [x] [[MAIN-68]] — github: contribution: invalid suggestion for edge case with zero
- [x] [[MAIN-69]] — boilerplate: fetch latest changes from go-tool before publishing v0.1.0
- [x] [[MAIN-70]] — github: contribution: refactor diff command
- [x] [[MAIN-71]] — pkg: file: register encoders for specific formats
- [x] [[MAIN-72]] — github: contribution: bad suggestion with Sunday shift
- [x] [[MAIN-73]] — github: contribution: suggest command has regression with weeks arg
- [x] [[MAIN-74]] — internal: pkg: remove golden package
- [x] [[MAIN-75]] — github: contribution: suggest shows future
- [x] [[MAIN-76]] — github: contribution: support strict ISO 8601 format as input date
- [ ] [[MAIN-77]] — github: contribution: snapshot and diff must support multiple date input
- [ ] [[MAIN-78]] — github: contribution: tips and tricks, daily snapshot
- [ ] [[MAIN-79]] — github: contribution: extend suggest command by stats
- [x] [[MAIN-80]] — github: labels: remove scope
- [ ] [[MAIN-81]] — makefile: build: allow include recursively
- [x] [[MAIN-82]] — deps: up go-github to v46.0.0
- [x] [[MAIN-83]] — deps: up go-github top v47.0.0
- [x] [[MAIN-84]] — github: contribution: invalid suggestion for past
- [x] [[MAIN-85]] — deps: up go-github to v47.1.0
- [x] [[MAIN-86]] — deps: up go-github to v48.0.0
- [x] [[MAIN-87]] — deps: up go-github to v48.2.0
- [ ] [[MAIN-88]] — git: stats: provide local-based statistics
- [x] [[MAIN-89]] — git: config: contribution since
- [x] [[MAIN-90]] — github: contribution: lookup call throw panic
- [ ] [[MAIN-91]] — github: cli: support extension mechanism
- [x] [[MAIN-93]] — github: contribution: invalid suggestion for 2022 year
- [x] [[MAIN-94]] — github: contribution: invalid suggestion for specific date
- [x] [[MAIN-95]] — ci/cd: fix broken build
- [x] [[MAIN-96]] — github: contribution: lookup command is corrupted
- [x] [[MAIN-97]] — github: contribution: suggest works incorrectly after refactoring
- [x] [[MAIN-98]] — github: contribution: suggest and lookup paniced on Sunday
- [x] [[MAIN-99]] — github: contribution: suggest use incorrect center
- [x] [[MAIN-100]] — github: contribution: incorrect shows stats for future
- [ ] [[MAIN-101]] — github: contribution: resolve all related to the scope todos
- [ ] [[MAIN-102]] — github: contribution: support preset for suggestion
- [ ] [[MAIN-103]] — github: contribution: calculate contrib to define better day for commits
- [ ] [[MAIN-104]] — github: contribution: refactor histogram command
- [x] [[MAIN-106]] — github: contribution: suggest for cmm
- [ ] [[MAIN-107]] — go: vanity: build from git submodules and go.work instead of modules.yml
- [ ] [[MAIN-108]] — go: rename: command to rename current module and others
- [x] [[MAIN-109]] — github: contribution: incorrect suggest for today
- [x] [[MAIN-110]] — github: contribution: edge case for contrib suggestion
- [x] [[MAIN-111]] — github: contribution: broken heatmap
- [ ] [[MAIN-112]] — template: go new support
- [x] [[MAIN-113]] — github: contribution: simplify arguments format
- [x] [[MAIN-114]] — github: up google/go-github to v55.0.0
- [x] [[MAIN-115]] — github: contribution: html markup was changed
- [x] [[MAIN-117]] — github: contribution: suggest doesn't work properly in headless mode
- [ ] [[MAIN-118]] — github: suggestion: reduce jitter
- [ ] [[MAIN-119]] — insights: deps: fetch data from deps.dev
- [x] [[MAIN-120]] — ci/cd: Continuous integration healthcheck doesn't work properly
- [x] [[MAIN-121]] — github: contributions: GitHub changed a way to provide data
- [x] [[MAIN-122]] — github: contribution: incorrect week number
- [x] [[MAIN-126]] — github: contribution: calendar counts are read from an inconsistent source
