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
- [ ] **The `review` skill is on trial** (since 2026-11-25): code, issue and
  security reviews of this repository are written with a local skill in
  `.agents/skills/review`, which is not tracked. The order is fixed: first
  prove it in use here; then make it part of the public agent skill in
  `skills/maintainer/`; then remove the review genre from the private
  `projects:maintainer` skill, which outside readers cannot see. Proven when
  its reports have carried a release, with their findings filed, fixed and
  shipped, and the owner confirms the benefit. If it does not prove itself,
  keep the genre of `projects:maintainer` and delete the local skill.
  The steps are filed in this order: octomation/skills-action#2 lets a
  tool publish a second skill; [MAIN-188](issues/MAIN-188.md) ships this one
  as `maintainer:review`; only after its release the private genre goes away,
  in its own repository.

- [ ] **Local issues [MAIN-164](issues/MAIN-164.md)–[MAIN-187](issues/MAIN-187.md)
  are not on GitHub**: they came from the full review of 2026-11-25 and wait
  for the owner to publish them. [MAIN-183](issues/MAIN-183.md)–[MAIN-187](issues/MAIN-187.md)
  and [MAIN-177](issues/MAIN-177.md) are the fixes of v0.5.1: publish them
  once the release is out, then tick their criteria and close them.
  [MAIN-164](issues/MAIN-164.md)–[MAIN-182](issues/MAIN-182.md) are planned
  for v0.5.2. The security review of 2026-11-25 is not published before
  v0.5.1 ships.
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
