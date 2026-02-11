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
- [ ] **On the first Nextra release after 4.6.1**: move the `nextra-theme-docs`
  override in `docs/package.json` to Zod 4.6.6 or later, or drop it, and check
  the static build. The Dependabot alert on Zod is dismissed as a tolerable risk
  until then ([docs/README.md](../docs/README.md)). Proven when the site builds
  and `docs/README.md` says so.
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

## Follow-ups to file

- [ ] **`status` is slow on one checkout**: on 2026-10-19 and 2026-10-20 the
  whole run over 164 checkouts took about 4 minutes, of which about 2.5 went
  to `git diff --find-renames --numstat HEAD` in `octolaba/changelog` (about
  2 s in September). File an issue with the timing per checkout.
- [ ] **`fetch` and an unconfirmed explicit repository**: the run fails as a
  whole with "cannot verify explicit checkout" when GitHub does not confirm the
  repository of an explicit rule; it should be a conflict of that repository.
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
