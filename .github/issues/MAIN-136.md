---
code: MAIN-136
id: null
databaseId: null
number: null
url: null
title: "fetch: apply: refresh successfully reconciled checkouts in the same run"
labels:
  - "type: improvement"
  - "scope: code"
  - "scope: test"
  - "scope: docs"
  - "impact: medium"
  - "effort: medium"
milestone: "[[MAIN-M3]]"
state: CLOSED
stateReason: COMPLETED
createdAt: 2026-09-26T09:28:00Z
updatedAt: 2026-09-26T09:28:00Z
lastEditedAt: null
closedAt: null
---

# fetch: apply: refresh successfully reconciled checkouts in the same run

**Motivation:** onboarding an existing tree takes two `--apply` runs — one to adopt, one to fetch — and the owner calls the second a superfluous step. The same gap follows every reconciliation of an existing checkout: after a `move`, a `relocate` or an `update_remote` the refs are refreshed only by the next invocation, so the transfer in the plan of 2026-09-26 would leave its checkout stale until the following week.

**Requirements**

- After a successful `adopt`, `move`, `relocate` or `update_remote` of an existing checkout, the same apply refreshes that checkout's refs once, last in the apply order and under the same concurrency bound. This is an extension of the existing ordering (`fetch` already sorts last), not a new mode or an action graph.
- A failed prerequisite prevents its dependent fetch. A successful adoption followed by a failed fetch is reported honestly — the record is adopted, the fetch error is counted, the exit code is `3` — and a rerun fetches without repeating the adoption.
- The bounded promise: after at most two successful applies against unchanged inputs, the following plan contains no executable reconciliation action — no adopt, relocate, move, update_remote or clone — only routine `fetch` and report-only orphan or noop rows. The planner adopts before it reconciles the canonical path and relocates before it reconciles path and origin, so a move or a remote update that follows an adoption is planned by the next apply; this issue does not chain adopt → move.
- A plan-only run shows the dependent fetch as part of the plan; JSON keeps listing every action.
- `fetch.md`, "Onboarding an existing tree", describes one apply and the second, idempotent one. If the change turns out to require rewriting the executor, the whole issue is deferred and the documentation keeps the second invocation.

**Evidence:** `internal/service/fetch/applier.go:40` executes one action per repository, `planner.go:396` plans `fetch` only for an unchanged record, `action.go` orders `fetch` last; the adoption-before-move order is at `planner.go:288` and the relocation at `planner.go:324`. The failure cases of a move followed by a failed remote update belong to [[MAIN-150]], and a failed state save to [[MAIN-137]].

**PoC**

A self-contained service test in `internal/service/fetch`, never the real collection, built like the existing lifecycle tests of `service_test.go`: a fake discoverer returning the snapshot `acme/tool`, a fake confirmer, a temporary root and an isolated state store, and the real Git adapter wrapped in `localOriginSync` (`service_test.go:46`), which reports the local origin as the canonical GitHub URL — the adopter skips non-GitHub origins (`adopter.go:105`), so a bare local remote is not enough on its own. The origin starts with an explicit commit on a branch named `main`, created by the test itself — the existing `originWithCommit` helper does not select that branch; `public/acme/tool` is cloned from it, then a second commit is pushed to the origin from another clone, so the checkout's `refs/remotes/origin/main` is stale. Then:

- one apply must report `adopt=1 fetch=1`, and `refs/remotes/origin/main` in the checkout must equal the new origin commit, while HEAD, the local `main` tip, the index, the working files and `remote.origin.pushurl` are byte-identical to before;
- a second apply must report `fetch=1` and nothing else;
- with the origin made unreadable between adoption and fetch, the first apply must record the adoption, count one error and exit `3`, and the next apply with the origin readable again must fetch without adopting.

Today the first apply reports `adopt=1 fetch=0` and the remote-tracking ref stays stale until the second run.

**Out of scope:** updating local branches or working trees; new action kinds; chaining a move after an adoption in one apply; retrying failed prerequisites automatically.

Related: [[MAIN-150]], [[MAIN-137]], [[MAIN-132]], [[MAIN-92]].

<!-- 2026-10-01T12:25Z
Preparation for implementation, from the spec reviews of #466 (2026-10-01): the dependent fetch is still absent on `main` — `service.go:475-479` runs fixed passes with only a failure counter (`service.go:435-440`), `applier.go:43-54` records an adoption without a fetch — and the user documentation says the opposite of the requirement: `docs/content/fetch.md:172` ("adoption/remote update fetch refs on the next invocation") and `:328` (a separate third command). The docs task must flip both lines when this lands. #466 deliberately does not build on this mechanism: `upstream` after clone or adopt is a second stage of the same action, like the origin update of `move`.
-->
