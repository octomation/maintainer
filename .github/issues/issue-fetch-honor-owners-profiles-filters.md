---
code:
id:
databaseId:
number:
url:
title: "fetch: selection: honor owners, profiles and adoption filters"
labels:
  - "type: bug"
  - "scope: code"
  - "scope: test"
  - "scope: docs"
  - "severity: major"
  - "impact: medium"
  - "effort: medium"
milestone: "[[milestone-3]]"
state: OPEN
stateReason:
createdAt: 2026-09-26T09:25:00Z
updatedAt: 2026-09-26T09:25:00Z
lastEditedAt:
closedAt:
---

# fetch: selection: honor owners, profiles and adoption filters

**Motivation:** `--owner`, `--profile`, `--token`, `include_owners` and `[filters]` are the only ways to say which repositories a run may touch. A narrowed run must not silently work across the whole collection or persist a false "not accessible" observation; mixed-case owners must not silently find nothing; filters must gate adoption as documented.

**Contract**: configured `include_owners` selects the discovery of new remote repositories and keeps the full-run behaviour of [the fetch guide](https://maintainer.octolab.org/fetch/), in which existing checkouts and pins outside those owners are still verified and managed. Explicit `--owner` and `--profile` narrow *this invocation*. The guarantee is about per-repository requests — confirmation by id, resolution by name, clone, fetch — and about mutation and persisted observations; a bulk listing may return repositories that are filtered out afterwards.

| Case | Per-repository requests | Mutation and fetch | Diagnostic | State |
| --- | --- | --- | --- | --- |
| tracked record of a selected owner; confirmation by id reveals it moved to an excluded owner | the confirmation is allowed: the record was selected | none — no move into an excluded owner's layout, no fetch | `noop` with the reason "transferred outside the selection" | unchanged |
| tracked record of an excluded owner; the listing of a selected owner shows the same id (transferred in) | no additional identity confirmation: the listing selected it; fetch is allowed after selection | allowed: move into the selected owner's path, remote update, fetch | the ordinary `move` line | updated on apply |
| id-only pin from the configuration with no local identity that decides eligibility | one confirmation by id: a pin is explicit user intent | by the confirmed owner; excluded → none | `noop` or `conflict` naming the reason | unchanged unless eligible |
| record of an excluded profile that a selected profile's listing also returns | allowed with the selected profile's credentials, provided the repository also passes any explicit `--owner` restriction | allowed with that profile's credentials and transport under the same owner restriction | ordinary | `source_profile` recorded only on a successful apply (§5.1 rule) |
| tracked record of an excluded owner, not listed by any selected profile | none | none | `noop` with the reason "outside the selection" in JSON; collapsed in human output; the summary shows selected of total | unchanged |

**Requirements**

- A 404 by id counts as "not accessible by id: deleted or access lost" only when it comes from the record's own profile token; with another token the record is left unverified, nothing is persisted, and the plan says why.
- Owners are compared case-insensitively everywhere: endpoint choice, memberships, allowlist, owner overrides, repo matches and ignore rules.
- `--token` applies after profile selection, to the first selected profile, never to every profile.
- Filtered snapshots are neither cloned nor adopted; tracked repositories stay tracked.
- Help, `fetch.md` and the specification say the same thing.

**Evidence:** findings 6 (`internal/service/fetch/service.go:402`), 8 (`internal/service/github/discovery_rest.go:122`, `internal/command/fetch/profiles.go:50`), 20 (`internal/service/fetch/planner.go:288` against `docs/content/fetch.md:81`) and 21 (`internal/command/fetch/root.go:56`, `service.go:176`, `docs/content/fetch.md:122` against `:125`) of [the code review](../reviews/20260926T071417Z.md), and G2 of [the review of 2026-07-10](../reviews/20260710T095546Z.md) (`internal/command/fetch/profiles.go:42`), reproduced on `932b9c8`.

**PoC**

A service test with a fake transport that records every request and answers case-insensitively like GitHub, an isolated state store seeded with records of owners `acme` and `other` and of profiles `primary` and `beta`, and a fake Git port:

- `include_owners = ["OctoCat", "Octomation"]` must produce the same endpoints and snapshots as the lower-case spelling; today the mixed-case spelling calls `/users/OctoCat/repos?type=owner` and `/orgs/Octomation/repos?type=public` and yields zero snapshots with `err = nil`;
- `--owner acme --apply` must leave the request log without any `/repositories/{id}` or fetch for `other/*`, leave their records byte-identical, and list them as `noop` with the reason in JSON; today they are confirmed and fetched;
- `--profile beta --apply` with a private record of `primary` must persist no `remote_status`; today it persists `"gone"`;
- `--profile beta --token T` must reach `beta`, not the profile at index 0;
- `exclude_forks = true` with an existing fork checkout must plan no `adopt`;
- `--profile beta --owner acme` together: a repository of `other` returned by `beta`'s listing gets no request, no mutation and no state change;
- the table above: one case per row.

**Out of scope:** GitHub App or device-flow authentication; a second profile for the owner's collection.

Related: [[issue-fetch-preserve-checkout-identity]], [[issue-fetch-state-honor-config-selection]], [[issue-docs-publish-fetch-and-status]].

<!-- 2026-10-01T12:25Z
Preparation for implementation, from the spec reviews of #466 (2026-10-01): reconfirmed on `main` — `planner.go:291-299` adopts before the filter check at `planner.go:304`, so `exclude_forks = true` still adopts a fork checkout. The fork-remote specification now also relies on this filter: with `exclude_forks = true` already tracked forks stay tracked, but `upstream` is neither configured nor checked. The onboarding config in `docs/content/fetch.md:316` sets `exclude_forks = true`, so fixing the order changes what a new user sees.
-->
