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
- [ ] **When `golang.org/x/crypto` tags a release after v0.58.0** (v0.4.1):
  bump it in `tools/` with `go get golang.org/x/crypto@<tag>` there. GO-2026-5932
  is fixed only in an untagged commit, so v0.4.1 shipped with `tools` failing on
  it; `cd` installs the tools but doesn't scan them. Proven when `tools` on
  `main` is green or fails only on advisories with `Fixed in: N/A`.
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

## Waiting for a decision

- [ ] **`cleanup.stale` is disabled** on purpose: `workflows` in
  `.github/settings.json` declares it, so `make doctor` reports it as `ok … on
  purpose`; drop that entry when enabling it. Once enabled, it marks every
  issue and pull request without an assignee, a milestone or `status: pending`
  after 30 days without activity, closes it 7 days later and deletes the
  branch of a closed pull request. Before `gh workflow enable
  cleanup.stale.yml`, list such open items and give the ones to keep an
  assignee, a milestone or `status: pending`.
- [ ] **Legacy make aliases** (`refresh`, `test-quick` and the other old
  names kept by the port of [MAIN-116](issues/MAIN-116.md)): drop them, and move
  every caller and issue text to the new names.
- [ ] **Open issues of released milestones**: "Hub, v0.3" (MAIN-25, MAIN-31,
  MAIN-105) and "Workspace, v0.2" (MAIN-19, MAIN-48, MAIN-116, MAIN-128) were
  not part of their releases. Move them to a next milestone, or keep the
  milestones open for patches.
- [ ] **Tool pins**: `make update` would move goreleaser in `tools/go.mod`
  ahead of the version the workflows pin. Hold `make update` until the pins
  move together; `make tool-pins-check` catches a drift.
- [ ] **`internal/command/proxy` and `internal/command/hub`** are not wired
  into the root command. Decide whether they have a purpose or go.

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
- [ ] **Texts of older issues** MAIN-11, MAIN-19, MAIN-25, MAIN-48 and
  MAIN-119 need refining to the current product. For MAIN-19: `UpdateRemote`
  does not check the expected previous URL, and `--apply` rebuilds the plan
  instead of applying the reviewed one.
- [ ] **Root `README.md`** on the layout of the documentation site, keeping
  the badges and every install channel.
- [ ] **Unused secrets** of the organization and the repository: review and
  remove the ones no workflow uses.
- [ ] **Deferred**: move from golang/mock to go.uber.org/mock; an agent quick
  start page.
