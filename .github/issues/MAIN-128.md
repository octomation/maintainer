---
code: MAIN-128
id: I_kwDOE2M9Zc8AAAABUVq71Q
databaseId: 5659868117
number: 466
url: https://github.com/octomation/maintainer/issues/466
title: "fetch: remotes: configure upstream for forks and check remote consistency"
issueType: null
labels:
  - "scope: code"
  - "scope: test"
  - "type: feature"
  - "impact: medium"
  - "effort: medium"
assignees:
  - kamilsk
milestone: null
parent: null
issueFields: {}
state: OPEN
stateReason: null
duplicateOf: null
createdAt: 2026-10-01T09:17:25Z
updatedAt: 2026-10-19T07:36:27Z
lastEditedAt: 2026-10-19T07:36:27Z
closedAt: null
---

# fetch: remotes: configure upstream for forks and check remote consistency

**Motivation:** fork checkouts should expose their parent as `upstream`, so that `git fetch upstream` and the `git please`/`git mirror` helpers work without a manual `git remote add`, and deviations from the remote convention should be visible in the plan before anything is applied. The convention: `origin` is the repository itself on GitHub, `upstream` is the repository it was forked from. This is the part of [[MAIN-92]] about fork relations that v0.2.0 does not ship.

**Evidence:** a manual audit on 2026-10-01 over `~/Development` reported 20 fork checkouts across `kamilsk` and six organisations (`octomation`, `octolaba`, `octopot`, `tact-app`, `lifeosm`, `withsparkle`), none with `upstream`. One of them, `kamilsk/homebrew-cask`, was cloned from the parent: `origin` was `Homebrew/homebrew-cask`, the fork was a remote named `fork`. Parents include the owner's own organisations (`octolab/.github`, `octolab/breaker`) and a parent under a renamed owner (`caarlos0-graveyard/watchub`).

**Requirements**

- Every remote change is a planned, reported action; nothing is rewritten silently.
- A fork checkout without `upstream` gets one, pointing at the current parent, read fresh on every run.
- A newly cloned or newly adopted fork gets `upstream` in the same apply. The first apply over an unknown tree only adds — state records and `upstream` — and rearranges nothing; moves and origin updates stay for the next apply.
- Identity comes from `origin`, never from other remotes. A path is a check, not a source of identity: a clone of the parent or of an unrelated repository at a fork's path is a conflict and is not adopted; a repository renamed after cloning is still adopted and moved by the next apply.
- An `upstream` that identifies a different repository, a changed parent, or several fetch URLs is a conflict reported next to the checkout's own action; it changes nothing and does not block the fetch of `origin`. A confirmed rename or transfer of the same parent may produce an explicit, reported fetch-URL update.
- A fork whose parent is gone and that is no longer a fork is reported as information, not as an error; its `upstream` is kept.
- A failure while configuring `upstream` never undoes a clone or an adoption; a rerun configures only `upstream`.
- With `exclude_forks = true`, `upstream` is neither configured nor checked.
- The push lock on `origin` survives, and `status` keeps showing the lock marker.

**PoC**

```bash
maintainer fetch --owner kamilsk --format=json | jq '.actions[] | select(.kind == "set_upstream")'
# expected: one action per fork that needs upstream, with remote, expected URL and reason
```

A checkout at the path of `kamilsk/homebrew-cask` whose `origin` is `Homebrew/homebrew-cask` appears in the plan as a remote-consistency conflict — not as an adoption of `Homebrew/homebrew-cask`, and not as an `upstream` addition.

**Out of scope:** mirror remotes on other hosts (an `upstream` outside GitHub is reported, not managed); pushing anywhere; repairing an inconsistent checkout automatically.

Related: [[MAIN-92]], [[MAIN-135]].

<!-- 2026-10-01T12:10Z https://github.com/octomation/maintainer/issues/466#issuecomment-5930768183
Preparation for implementation: findings of two spec reviews (Fable 5.1 and gpt-6-astra, high effort), checked against `main`. Nothing here is fixed yet.

- **The parent clone is adopted today.** `planner.go:288` adopts a workspace checkout without a path or target cross-check, and `service.go:176-208` confirms any workspace checkout by id, so the inverted `homebrew-cask` case becomes `adopt Homebrew/homebrew-cask` and a `move` on the next run. The PoC of this issue fails on `main`.
- **No remote but `origin` is visible.** `CloneInfo` carries only origin fetch and push URLs (`internal/service/git/contract.go:32-36`), read from the `origin` subsection only (`sync.go:187-198`); `adopter.go:105` uses `Origins[0]`. `Inspect` has to return all remotes.
- **No parent anywhere.** `RepoSnapshot` (`internal/service/github/contract.go:33-49`), `mapRepo` (`discovery_rest.go:264-273`) and `state.Record` (`internal/state/state.go:20-42`) have no parent fields. The parent read belongs to the service between the scan (`service.go:150`) and `Plan` (`service.go:225`); the planner stays pure.
- **`upstream` after clone or adopt is a second stage, not a dependent action.** The only shipped two-stage precedent is `move` + `UpdateRemote` (`action.go:55`, `applier.go:79-97`, reporter sub-line `reporter.go:176-178`). The planner emits one action per id (`planner.go:164-172`) and the executor runs fixed passes without per-id outcomes (`service.go:435-440`, `:475-479`); the dependent fetch belongs to the v0.2.0 task "fetch: refresh reconciled checkouts" and is not a prerequisite here.
- **Extra rows per id.** `upstream` conflicts and the `upstream-unknown` / `upstream-detached` flags are rows next to the checkout's action, like orphan copies; today one action per id would make them replace the fetch. `reporter.go:263-272` (`actionable`) must count `set_upstream`, otherwise a plan with only that action prints "no drift".
- **Do not copy `UpdateRemote` as is.** It reports success without re-reading the push configuration (`sync.go:160`); the `upstream` change must verify both fetch and push URLs after re-reading. The origin-only refspec constant at `sync.go:19` needs a per-remote variant.
- **JSON.** Existing fields are omitted when empty (`reporter.go:282-316`); the new `remote`, `actual_urls`, `expected_url` and `set_upstream` follow that.
-->
