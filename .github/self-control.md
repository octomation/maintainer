# Self-control

Open checks that outlive a working session: things to look at later, decisions
that are waiting, and follow-ups that still need an issue. Each item says when
to look, what proves it and what to do if it fails. An item is removed once it
is proven or moved into an issue; a release adds the checks it leaves open
([AGENTS.md](../AGENTS.md#7-self-control-and-cleanup)).

## Watch

- [ ] **Until 2027-01-07** ([MAIN-159](issues/MAIN-159.md), v0.4.0): `lookup`
  agrees with the profile on the edges of months, on reviews after 15:00 UTC
  and on 2026-12-31 and 2027-01-01:
  [contribution-profile-diff.mjs](scripts/contribution-profile-diff.mjs) checks
  a snapshot of the year against the profile. A day that differs goes into
  [MAIN-160](issues/MAIN-160.md) with its date and both counts.
- [ ] **On the first Nextra release after 4.6.1** (shuding/nextra#4990): move
  the `nextra-theme-docs` override in `docs/package.json` to zod 4.6.6 or
  later, or drop it, and drop the katex and minimatch overrides its own
  ranges cover. Proven when `npm audit` in `docs/` is clean and the site
  builds.

Overrides in `docs/package.json`:

| Package | Override | Why |
| --- | --- | --- |
| katex | `^0.18.2` | advisories below 0.18.2; Nextra and mermaid ask for 0.16 |
| minimatch | `^11.0.1` | advisories below 11.0.1; Nextra → ts-morph asks for 10 |
| zod | `4.1.12` under `nextra-theme-docs` | **kept vulnerable on purpose**: newer zod breaks Nextra 4.6.1 (shuding/nextra#5036). It validates the repository's own config at build time only; the alert (GHSA-mrhx-c7xc-m3w5) is dismissed as a tolerable risk |
| @xmldom/xmldom | `^0.9.12` under `speech-rule-engine@4.1.4` | advisories of the version speech-rule-engine asks for |

- [ ] **After the next reboot of a Mac with the service** (v0.3.0): `brew
  services` starts the hub by itself: `launchctl list | grep sh.brew.maintainer`
  has a PID and `/status` answers `200` before anyone restarts it. On
  2026-11-07 the log had six starts and no errors, but every start since the
  upgrade was by hand.
- [ ] **Next time the hub is open on the real collection** (v0.3.0,
  [MAIN-155](issues/MAIN-155.md)–[MAIN-158](issues/MAIN-158.md)): `/status`
  shows every row of `maintainer status --format=plain` and sorts and searches
  without a visible delay; setting `impact: high` in `/inbox` changes only that
  label line of the issue file (`git diff`); `/tablo` shows the columns of
  `MAIN-P1` and the `MAIN-M3` filter adds up to the milestone's count. Tick the
  matching criteria of those issues, or file what differs.

- [ ] **After the upgrade to v0.5.1 on a Mac with the service**
  ([MAIN-184](issues/MAIN-184.md)): with `access = "open"` in
  `service.toml`, an open tab of `http://maintainer.localhost:17841/` shows
  the data after `brew services restart octolab/tap/maintainer` without any
  command, and the theme stays. Once, `maintainer service open` opens the
  label address with data in the browser in daily use (Safari, Chrome,
  Firefox), and `ls ~/.local/state/maintainer/service/` shows `17841.label`
  with mode `0600` while the service runs. The label flow was checked in
  headless Chrome only: a browser that does not open the label address goes
  into an issue with its name and version.

## Waiting for a decision

- [ ] **`cleanup.stale` is disabled** on purpose: `workflows` in
  `.github/settings.json` declares it, so `make doctor` reports it as `ok … on
  purpose`; drop that entry when enabling it. Once enabled, it marks every
  issue and pull request without an assignee, a milestone or `status: pending`
  after 30 days without activity, closes it 7 days later and deletes the
  branch of a closed pull request. Before `gh workflow enable
  cleanup.stale.yml`, list such open items and give the ones to keep an
  assignee, a milestone or `status: pending`.
- [ ] **Tool pins**: `make update` would move goreleaser in `tools/go.mod`
  ahead of the version the workflows pin. Hold `make update` until the pins
  move together; `make tool-pins-check` catches a drift.

- [ ] **Local issues [MAIN-166](issues/MAIN-166.md),
  [MAIN-168](issues/MAIN-168.md), [MAIN-170](issues/MAIN-170.md),
  [MAIN-182](issues/MAIN-182.md), [MAIN-189](issues/MAIN-189.md) and
  [MAIN-190](issues/MAIN-190.md) are not on GitHub**: the rest of the full
  review of 2026-11-25, planned for v0.6.1. Publish them before their code
  changes.
- [ ] **The hub speaks plain HTTP** (v0.5.1): a program of another account
  that takes the port while the service is down sees the addresses that an
  open tab still requests, and can show a false page. The service guide states
  this limit. HTTPS with a local certificate would close it and is not
  planned; decide again if the hub gets users on shared computers.

## Follow-ups to file

- [ ] **A rename race in the Git adapter**
  (`internal/service/git/sync.go`): go-git writes by path after the
  reservation, and `Move` checks with `Lstat` before `Rename` without a
  no-replace rename.
- [ ] **Texts of older issues** MAIN-19, MAIN-25, MAIN-48 and MAIN-119 need
  refining to the current product. For MAIN-19: `UpdateRemote` does not check
  the expected previous URL, and `--apply` rebuilds the plan instead of
  applying the reviewed one.
- [ ] **Root `README.md`** on the layout of the documentation site, keeping
  the badges and every install channel.
- [ ] **Unused secrets** of the organization and the repository: review and
  remove the ones no workflow uses.
- [ ] **Deferred**: an agent quick start page.
- [ ] **Hub tests that the criteria still miss**: a `/status` component test
  against a fixture recorded with a sort chain and a query
  ([MAIN-156](issues/MAIN-156.md)), and the 3000-task check of
  [MAIN-155](issues/MAIN-155.md) through `GET /tasks` over HTTP rather than in
  process.

## Checks by occasion

Not open work: what to check each time the occasion comes.

### After Dependabot fails or a security alert opens

A security update of a transitive dependency of `docs/` fails with
`security_update_not_possible` when its parent pins an older range.

- [ ] **Lift it with `overrides`** in `docs/package.json`, then `npm install`
  and `npm audit` in `docs/`, and build the site. Proven when `npm ls <name>`
  shows the patched version and the site builds; add a row to the table of
  overrides above.
- [ ] **When an override cannot be lifted** (the site breaks), dismiss the
  alert as a tolerable risk with the reason and the upstream issue, and say
  so in the table.
