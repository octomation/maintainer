# GitHub Actions

| Workflow | Name | Runs on | Does |
| --- | --- | --- | --- |
| [ci](#ci) | Continuous integration | PR and push to main (Go files, release config), `v*` tag, monthly, manual | Lint, tests, coverage to Codecov, vulnerabilities, release config |
| [cd](#cd) | Continuous delivery | `v*` tag, manual | Check the tag, test, publish the release, the Homebrew Formula and Cask; a snapshot on manual runs |
| [docs](#docs) | Documentation delivery | PR and push to main (`docs/`), monthly, manual, reusable | Build the site; deploy it to Pages from main |
| [tools](#tools) | Tools validation | PR and push to main (`tools/`), monthly, manual | Install the tools module, check generated code and tool vulnerabilities |
| [doctor](#doctor) | Repository doctor | daily, manual | Compare the repository with GitHub and explain fixes |
| [caches](#caches) | Workflow caches cleanup | monthly, manual, reusable | Delete all Actions caches |
| [warmup](#warmup) | Workflow caches warmup | after caches cleanup, push to main (`go.sum`, `tools/go.sum`, `.goreleaser.yml`), twice a week, manual | Refill Go, docs and tools caches; keep the release build cache warm |
| [pulls](#pulls) | Pull request caches cleanup | PR closed, daily, manual | Delete the caches of closed PRs |
| [runs](#runs) | Workflow runs cleanup | monthly on day 2, manual, reusable | Delete completed runs, with a dry run |
| [stale](#stale) | Stale issues cleanup | daily, manual, reusable | Mark and close stale issues and PRs |

Schedules are in UTC: cleanups and the doctor at 06:00, checks at 07:00, monthly on day 1;
runs cleanup on day 2, so the runs of day 1 stay readable until then; the release warmup on Mondays and Thursdays.

```mermaid
flowchart LR
  dev([developer]) -- PR / push main --> ci & docs & tools
  dev -- "git push --atomic origin main vX.Y.Z" --> hook{{pre-push hook}}
  hook -- tag --> cd & ci
  cron([schedule]) --> ci & docs & tools & doctor & caches & runs & stale & warmup & pulls
  caches -- completed --> warmup
  dev -- PR closed --> pulls
  cd --> release[(GitHub release)] & tap[(Homebrew tap)]
  docs --> pages[(GitHub Pages)]
```

## Common ground

- All external actions use concrete release tags; Dependabot checks the
  `github-actions` ecosystem daily. Review upstream changes and the action
  contract before accepting an update.
- GitHub-hosted Ubuntu runners (`ubuntu-24.04`) supply the runtime the actions need.
- Go follows the current stable release: `go.mod`, `tools/go.mod`, `GO_VERSIONS`
  in the `Makefile` and every `setup-go` move together; tests also run on `1.x`.
- Workflow permissions default to `contents: read`; publishing and maintenance
  jobs request more explicitly.
- Every workflow except doctor and pulls ends with a `notify` job that posts to Slack.
  `SLACK_WEBHOOK` is optional: an empty value skips the notification. Commit
  messages and dispatch reasons reach the message through environment variables.
- Secrets are checked by `make doctor`, which prints how to create a missing one.

| Secret | Used by | Purpose |
| --- | --- | --- |
| `OCTOLAB_RELEASER_CLIENT_ID` | cd, doctor | Client ID of the GitHub App "OctoLab Releaser" that pushes the Formula and the Cask to the tap named in `.goreleaser.yml` |
| `OCTOLAB_RELEASER_KEY` | cd, doctor | Private key of that App |
| `SLACK_WEBHOOK` | all but doctor and pulls | Notifications, optional |

### Set up GitHub

Nothing here is stored in the repository, so a fresh repository or a fork needs it once:

1. Settings → Pages → Source: **GitHub Actions**; the `github-pages` environment
   must allow deployments from main. Custom domain: `maintainer.octolab.org`,
   declared as `pages.cname` in `.github/settings.json`, with Enforce HTTPS on.
2. `OCTOLAB_RELEASER_CLIENT_ID` and `OCTOLAB_RELEASER_KEY`: the Client ID and a
   private key of the GitHub App "OctoLab Releaser", owned by `octolab`, with
   no webhook and Repository permissions → Contents: Read and write, installed
   on `octolab/homebrew-tap`. Save both as secrets of the `octomation`
   organization available to this repository. The release itself uses the
   workflow token.
3. Branch protection of main: require signed commits; `.github/settings.yml`
   cannot set it.
4. Activate the repository in Codecov; the upload uses OIDC, no `CODECOV_TOKEN`.
5. Enable the workflows that were disabled by hand, after pushing the YAML that
   should run: `gh workflow enable <file> -R octomation/maintainer`. The doctor
   reports every workflow that is not active, except those that
   `workflows` in `.github/settings.json` keeps off on purpose:
   `cleanup.stale.yml` for now.

Then dispatch ci, tools and docs, run doctor after the first Pages
deployment (`make workflow-doctor`), and dispatch cd on main for a snapshot.
Do not dispatch the cleanups to try them out: they delete data.

## ci

[ci.yml](ci.yml) keeps the main branch green and reports coverage.

```mermaid
flowchart LR
  lint[lint: golangci-lint]
  test[test: vet and format, tidy modules, tests, coverage, install check] --> report[report: Codecov via OIDC]
  vulns[vulns: tool pins, release helpers, settings, goreleaser config, application vulnerabilities, generated code]
  lint & test & report & vulns --> notify[notify]
```

- Runs on PRs and pushes to main that touch Go code (outside `tools/`),
  `go.mod`, `Makefile`, `Taskfile`, the lint or release configuration, the
  release helpers or hooks, or the workflow itself; on `v*` tags; monthly; manually.
- Every gate of the release runs here first: an untidy `go.mod`, a broken
  `.goreleaser.yml` or a vulnerable dependency fails on main, not on the tag.
- `vulns` scans the application only; tool vulnerabilities belong to [tools](#tools).
- Codecov uses GitHub OIDC (`id-token: write`). The repository must be activated
  in Codecov; no `CODECOV_TOKEN` is needed.
- PRs get no notification.

## cd

[cd.yml](cd.yml) publishes a release from a tag.

```mermaid
flowchart TB
  tag([push v* tag]) --> check[check the tag and its note] --> signature[GitHub verifies the tag] --> secrets[check release secrets] --> token[mint the tap token] --> pages[resolve the Pages URL] --> render[render the note]
  render --> go[set up Go, restore the release build cache, make tools] --> test[release-verify and tests] --> publish[goreleaser release]
  publish --> release[(GitHub release: body and title from the note)]
  publish --> tap[(Formula and Cask in the tap)]
  publish --> notify[notify]
```

- **A release is a curated note plus a tag.** Write `docs/content/changelog/<tag>.md`:
  frontmatter `title` and `description`, the first line `# <title>`, Markdown
  and HTML, no MDX, site-relative links. Commit it, tag the commit with a
  signed tag, run `make release-check TAG=<tag>`, then push the branch and the
  tag in one go: `git push --atomic origin main <tag>`.
- **Guardrails come first and fail in seconds.** `make release-check` checks the
  tag against the local branch, before it is pushed, together with tidy modules
  and the GoReleaser config. The `pre-push` hook (`.github/hooks`, wired by
  `make git-hooks`, part of `make setup`) runs `.github/scripts/release.mjs check`
  against the branch of the same push, then the application checks of every
  push, `make go-deps-tidy verify`. The first steps here repeat the tag check
  against `origin` for pushes that bypassed the hook, check that GitHub
  verifies the tag's signature (`release.mjs signature`), then mint the tap
  token, all before Go is even installed.
- **The note becomes the release.** `release.mjs render` strips the frontmatter
  and the H1, makes site links absolute from the Pages URL and hands the title
  to goreleaser (`release.name_template`).
- **Policy lives in `.github/settings.json`**, documented by `.github/settings.cue`:
  tag pattern, maintenance branches (e.g. `v5.*` tags on branch `v5`), note path.
- The release uses the workflow token. The Formula and the Cask go to the tap
  named in `.goreleaser.yml` with a token minted per run by
  `actions/create-github-app-token` from a GitHub App installed on the tap,
  with Contents: Read and write; minting it fails if the App cannot write
  there. goreleaser then omits the committer
  (`commit_author.use_github_app_token`), so GitHub signs the tap commits as
  the App; a personal token does not get signed commits.
- The Formula and the Cask install the same release on macOS and Linux, see
  [MAIN-152](../issues/MAIN-152.md). Homebrew
  cannot declare a conflict between a formula and a cask, so the Formula
  carries a caveat and the Cask preflight warns when the Formula owns the
  binary link. GoReleaser deprecates `brews`, so `make release-config-check`
  accepts exactly that deprecation and fails on any other.
- A tag with a prerelease suffix, such as `v0.3.0-rc1`, is published as a
  prerelease that never becomes *Latest*: with `release.prerelease: auto`
  GoReleaser marks it so and sends `make_latest: false`. `skip_upload: auto`
  on `brews` and `homebrew_casks` leaves the Formula and the Cask in the tap
  as they are, see [#491](https://github.com/octomation/maintainer/issues/491).
- The binary is neither signed nor notarized. The Cask preflight clears its
  quarantine attribute: it has to run before the completions are generated,
  which execute the binary, so a `postflight` hook would be too late.
- A manual run on a branch builds a snapshot and publishes nothing; on a tag it
  publishes, as a tag push does.

## docs

[cd.docs.yml](cd.docs.yml) builds the Nextra site and deploys it to GitHub Pages.

```mermaid
flowchart LR
  check[Pages URL matches settings.json pages] --> build[build: npm ci, static export with BASE_PATH and SITE_URL] --> deploy{main, not a PR?}
  deploy -- yes --> pages[(GitHub Pages)] --> smoke[smoke test: pages, assets, og:image, 404, redirect]
  deploy -- no --> skip([artifact only])
  build & pages --> notify[notify]
```

- Runs on PRs and pushes to main that touch `docs/`, `.github/settings.json`,
  `release.mjs`, `Taskfile`, `bin/lib/` or the notify action; monthly;
  manually; as a reusable workflow.
- On main, `BASE_PATH` and `SITE_URL` come from the Pages configuration; PR
  previews take them from `.github/settings.json` (`release.mjs site-env`)
  and need no Pages access.
- No `CNAME` file is needed: a custom domain is set in Settings → Pages and
  declared as `pages.cname` in `.github/settings.json`; the build stops while
  they disagree. The domain is baked into the build, so rebuild after changing
  it (see [docs/README.md](../../docs/README.md)): switch Settings → Pages
  first, then merge the new `pages.cname`; the site serves stale asset paths
  until that deploy finishes. A fork deploying its own Pages must drop or
  change `pages.cname`.
- Independent of releases: merge a note and its new pages before tagging if the
  release must link to them from the first minute.

## tools

[tools.yml](tools.yml) keeps the tools module installable and its generated code in sync.

```mermaid
flowchart LR
  install[make tools, one binary per tool directive] --> tidy[go-tools-tidy, generate, git-check] --> pins[tool-pins-check] --> check[go-tools-check: verify, govulncheck tool] --> notify[notify]
```

- Runs on PRs and pushes to main that touch `tools/**.go`, `tools/go.mod`,
  `tools/go.sum`, the `Makefile` or the workflow itself; monthly; manually.
- `make tool-pins-check` catches what `make go-tools-update` leaves behind:
  the golangci-lint and goreleaser versions that ci, cd and warmup pass to
  their actions must be the ones `tools/go.mod` requires; the Go versions in
  `tools/go.mod`, the workflows, `GO_VERSIONS` of the `Makefile` and "Go X.Y
  or newer" in the docs must be on the Go line of `go.mod`; the workflows
  must agree on Node.js. It runs here before the scan, in ci `vulns` and in
  the doctor.
- The vulnerability scan of the tools fails only this workflow, never CI or a
  release; see [tools/README.md](../../tools/README.md).

## doctor

[doctor.yml](doctor.yml) runs `release.mjs doctor` and `preflight` in CI; `make doctor` runs the first one locally, and `make workflow-doctor` dispatches the workflow on main and waits for its result.

```mermaid
flowchart LR
  doctor[compare with GitHub: default branch, workflows, Pages, site smoke test, go.octolab.org imports, secrets, tool pins, goreleaser config] --> preflight[check release secrets] --> token[mint the tap token]
```

- Daily and manual: a Pages domain change triggers no workflow, so a site
  built for the old domain is caught here. Each problem is printed with what
  to fix and where.
- `go.octolab.org` is in `GOPRIVATE`, so `go` resolves it directly: every
  vanity module in `go.mod` and `tools/go.mod` must answer `?go-get=1` over
  verified HTTPS with a `go-import` tag. A lapsed certificate fails here before
  a fresh runner hits it; `node .github/scripts/release.mjs vanity` checks
  only that.
- The workflow token cannot list secrets, so they show as `unverified` there;
  the preflight step checks the ones a release needs, and minting the tap
  token proves the App is still installed on the tap with write access.
  Locally, `gh` sees the organization secrets shared with the repository.

## caches

[cleanup.caches.yml](cleanup.caches.yml) deletes every Actions cache with the built-in GitHub CLI.

```mermaid
flowchart LR
  delete["gh cache delete --all"] --> notify[notify]
  delete -. completed .-> warmup([warmup])
```

## warmup

[warmup.caches.yml](warmup.caches.yml) refills the caches right after a successful cleanup.

```mermaid
flowchart LR
  deps[Go modules] & docs[docs npm] & tools[tools module] & release[release builds] --> notify[notify]
```

- The keys match the consumers: the Go matrix of ci uses `go.sum`; the tools
  job uses `tools/go.sum`; ci `vulns` uses both.
- **The release build cache lives on main.** A tag can restore only caches of
  its own ref and of the default branch, so the cache `cd` restores is made
  here: `goreleaser build` of every target, then the tools and the tests that
  `cd` runs, keyed by the Go version and `go.sum`, `tools/go.sum`,
  `.goreleaser.yml` and the warmup itself. A push to main that changes one of
  them rebuilds it, incrementally from the previous one, a run twice a
  week keeps it inside GitHub's 7-day eviction window, and the other jobs run
  only after a successful cleanup or on demand.
- `cd` builds with full parallelism only on the exact key; any other match
  brings the modules but not a warm build cache, so it keeps `--parallelism 2`.

## pulls

[cleanup.pulls.yml](cleanup.pulls.yml) deletes the caches of a pull request once
it is closed: they are scoped to `refs/pull/<n>/merge` and of no use after it,
but count against the 10GB quota. It runs on `pull_request_target`, which keeps
a writable token when Dependabot closes its own PR, and checks out no code.
A run still in progress when its PR closes saves its caches afterwards, so a
daily run sweeps the caches of every PR that is no longer open.
It has no `notify` job, as it runs on every closed PR.

## runs

[cleanup.runs.yml](cleanup.runs.yml) deletes completed workflow runs, without age or count retention.

```mermaid
flowchart LR
  delete[delete completed runs] --> notify[notify]
```

- To clear the history, run it with `pattern: All` (the default) and `dry_run`
  unchecked. Scheduled and reusable runs also target all workflows, Dependabot
  included. Active runs, including this one, remain.
- The upstream action also deletes orphaned runs of workflows that no longer
  exist, whatever their status or the selected workflow.

## stale

[cleanup.stale.yml](cleanup.stale.yml) marks inactive issues and PRs, then closes them.

```mermaid
flowchart LR
  check[mark and close stale issues and PRs] --> notify[notify]
```

- Issues and PRs with an assignee, a milestone or `status: pending` are exempt.
- [.github/integrations/stale/stale.yml](../integrations/stale/stale.yml) keeps
  the settings of the retired Stale app for reference; nothing reads it.

## Local checks

```sh
make tools
make source-check lint test
make go-deps-tidy go-tools-tidy generate git-check
make config-vet release-config-check go-deps-check
make go-tools-check
node --test .github/scripts/*.test.mjs
actionlint
make go-dist-check
```
