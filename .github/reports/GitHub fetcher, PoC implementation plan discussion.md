---
uid: 67c4ef68-aa26-4208-8eaa-df4f63fe8658
completed: 2026-06-05
confirmed: 2026-06-05
---
# GitHub fetcher — PoC plan review and consensus

A record of a design review of the `maintainer fetch` PoC implementation plan and the
consensus reached on concrete changes to that plan. Two independent reviews — referred to
here as **C** and **G** — were consolidated into a single set of findings, after which the
reviewers exchanged positions until they agreed on a set of specification edits.

Code paths (`internal/...`) describe the project layout and are part of the decisions;
everything else is environment-independent.

## Outcome

Both reviews agreed the plan is strong as an architectural design (scope, non-goals, CLI
surface, state model, planner/apply model, milestones, test strategy) but **not yet
implementation-ready**: several statements about the existing codebase were inaccurate, and a
number of decisions that affect the first milestones were left open — Git authentication for
clone/fetch, path and state semantics, exit codes, locking, and adopter edge cases. The
review converged on **32 resolved findings**, **3 negotiated trade-offs**, and **14 concrete
specification changes** (listed at the end).

Each finding below is recorded as **Issue → Resolution**, grouped by theme. Where the two
reviews genuinely disagreed, the negotiated outcome is shown separately with both stances.

## 1. Baseline accuracy

These block "implementation-ready" because the "reuse as-is" section depends on them.

1. **Ports do not already exist.** The plan claimed a `Discoverer`/`GitSync` contract file
   "already exists"; it does not. **Resolution:** the ports are new. Split them by domain —
   `Discoverer` in `internal/service/github`, `GitSync` in `internal/service/git`, `StateStore`
   alongside the state package — so the GitHub package does not own non-GitHub contracts. Mocks
   via the existing `mockgen` setup.

2. **An existing Git service is unmentioned.** `internal/service/git` already wraps go-git
   behind a small interface. **Resolution:** do not introduce a parallel `gitsync` package;
   extend `internal/service/git`, whose role changes from a remote-inspection wrapper to the
   `GitSync` port. No public package is needed.

3. **`afero` does not back go-git.** The plan promised filesystem-free `GitSync` tests on an
   in-memory `afero`; go-git works over a `billy` filesystem or an OS path, not `afero`.
   **Resolution:** keep `afero` for config/state I/O only. Test `GitSync` with go-git's
   in-memory/billy filesystem (pure tests for planner/state) plus temp-directory integration
   tests for `Move`, permissions, and collisions.

4. **Package placement.** The plan proposed a public top-level package tree that conflicts with
   the project's internal-only convention. **Resolution:** keep everything internal. Fetch
   config lives next to the existing config package (one config package, not two); the state
   store under an internal state package.

5. **Logging.** The plan both claimed to "standardise on `slog`" and to "match the existing
   convention", which is actually cobra's command printers. **Resolution:** for the core
   milestones, observability is the Reporter (stdout: plan and summary) plus the command's
   stderr writer (progress and errors). `slog` is introduced only at the polish milestone for
   leveled `-vv`/`-vvv` logging, framed as a deliberate new addition — not an existing
   convention.

6. **Token binding.** The token/flag binding lives in the sibling command group, not in the
   shared config type, so the new command does not inherit it. **Resolution:** extract a small
   shared helper that binds the GitHub token env var and `--token`, called by both command
   groups. The unrelated remote flag is not pulled into the new command (no consumer).

7. **Bounded concurrency.** The cited prior use of `errgroup` is unbounded. **Resolution:** use
   `errgroup.SetLimit(concurrency)`; validate `concurrency >= 1`; the limit is set before any
   goroutine starts and is not changed within a phase.

8. **Exit codes.** The entrypoint currently collapses every error to exit `1`, but the plan
   specifies `1/2/3`. **Resolution:** introduce a typed error carrying an exit code; the
   entrypoint maps it (default `1`, user/config error `2`, partial apply `3`). This is a
   first-milestone task so the CLI contract is testable from the start.

## 2. Authentication and access

The largest gap: the plan covered the REST PAT but not credentials for Git operations.

9. **Git transport auth.** How clone/fetch authenticate for private repositories was
   unspecified. **Resolution (negotiated — see trade-off 1):** HTTPS + PAT is the supported,
   tested path for private repos; SSH is best-effort via an already-configured agent and strict
   host-key checking, with no key/passphrase management and no host-key file mutation.
   Per-operation auth, never embedded in a URL.

10. **Whose credentials on overlap.** When a repo is visible from two profiles, the snapshot
    selection rule picks the winner, but the credential source was undefined. **Resolution:** the
    winning snapshot carries its source profile; clone/fetch use that profile's credentials.

11. **Source profile in state.** **Resolution:** store the source profile (no secrets) on the
    record and action. A scope filter that excludes a profile must not silently fetch its repos
    via the stored profile; falling back to the stored profile is allowed only in a full run
    where that profile is present and its token resolves. A missing token is a user/config error,
    not a silent switch to another token.

12. **Credential leakage into stored URLs.** **Resolution:** the stored and displayed remote URL
    is the canonical, credential-free URL; auth is supplied per operation. At adoption, a remote
    that contains an embedded token is normalised to the redacted canonical URL in state, and the
    plan reports auth/transport drift without printing the secret.

## 3. Configuration, paths, flags

13. **No config file present.** **Resolution:** no config and no owner flag → exit `2`
    ("not configured"). No config but an owner flag → single-run mode with a default profile from
    the token env/flag. A token is required even for public-only discovery, to avoid two
    rate-limit/error regimes.

14. **Path resolution rule.** **Resolution:** a rendered template is used as-is if absolute,
    expanded against the home directory if it starts with `~`, otherwise joined onto the root and
    cleaned. Default and per-owner templates must stay within the root (a `..` escape is an
    error); absolute and `~` paths are allowed only for per-repo overrides.

15. **Paths outside the root.** **Resolution (negotiated — see trade-off 2):** the disk scan
    covers the root plus the explicit override paths from config. External override paths are
    adopt/fetch-only and are never auto-moved in the PoC; a changed override is reported as drift,
    not executed as a relocation of a user directory.

16. **Plan-only vs apply flags.** A default-true dry-run flag alongside an apply flag is
    awkward. **Resolution:** drop the dry-run flag from the surface; absence of `--apply` is
    plan-only. If retained at all it is a hidden no-op, and combining it with `--apply` is exit
    `2`.

17. **Config-init overwrite.** **Resolution:** "no prompts" does not mean silent overwrite. An
    existing file → exit `2`; a `--force` flag overwrites. The created-or-refused path is printed.

18. **Filters vs tracked state.** **Resolution:** filters gate only new clone/adopt decisions.
    Discovery still collects everything from declared owners, so a tracked repo that newly matches
    an exclude filter stays tracked and becomes a no-op/fetch flagged "filtered" — never an orphan
    or prune candidate. Filters are never applied as API-level exclusion.

19. **Internal visibility.** The schema allowed an internal visibility the path variable could
    never produce. **Resolution:** source visibility from the API visibility field (falling back
    to the private boolean only when empty); the enum and default tree include the internal value.

## 4. Planner / apply / collisions

20. **Collision taxonomy.** **Resolution:** an explicit table for the target-path state at
    clone time — empty/absent → clone; a Git repo whose origin matches the same id → adopt; a Git
    repo with a different id, no origin, or multiple origins → conflict; a file or a non-empty
    non-Git directory → conflict; a Git-file/worktree or bare repo → conflict (out of PoC scope).

21. **Move onto an occupied path / cross-device.** **Resolution:** never overwrite (occupied
    target → conflict). The PoC performs only same-volume rename; a cross-device move fails the
    repo with a clear message (exit `3` under apply). No copy+remove fallback (partial-copy
    corruption risk).

22. **Collisions within a single plan.** **Resolution:** the planner builds a global target-path
    index before apply; two actions resolving to the same canonical path are a conflict for both
    repos, decided before any side effect — not a matter of execution order.

23. **Fetch semantics.** Two concerns pulled opposite ways (see trade-off — resolved here).
    **Resolution:** behaviour is unchanged — fetch runs for all tracked repos and only updates
    remote-tracking refs (the working tree and local branches are never touched; this is recorded
    as an explicit non-goal). The human plan collapses routine fetches into a summary count and
    prints only drift; the JSON plan lists routine fetches in full so it remains a complete record
    of side effects.

24. **Update-remote trigger.** **Resolution:** update-remote fires only when the canonical URL
    for a tracked repo's observed transport differs from the actual origin URL (i.e. owner/name
    changed). A change of the configured transport preference does not rewrite existing remotes;
    new clones use the current config, existing remotes keep their transport until an explicit
    future normalise command.

25. **Adoption after rename.** Matching a disk clone to a snapshot by owner/name fails if the
    repo was renamed after cloning. **Resolution:** on a name miss, resolve the on-disk
    owner/name through the API (following the rename redirect) to recover the stable id, then
    match by id. If the old name now belongs to a different repo, or the redirect does not resolve
    into the current snapshot set, it is a conflict — not a guess.

26. **Manual relocation.** **Resolution:** the disk scan indexes every discovered clone by
    resolved id, including already-tracked ones. If a record's path is missing but the same id is
    found at exactly one other location, the action is a state relocation, not a re-clone. The
    same id found in multiple locations is an ambiguity conflict.

## 5. Discovery / rate limits / orphan

27. **Per-profile de-duplication.** The same repo appears via the user endpoint and an org
    endpoint. **Resolution:** de-duplicate by id within a profile before the cross-profile merge;
    the record with broader visibility / fuller fields wins, otherwise a deterministic
    endpoint-priority tie-break.

28. **Owner allowlist.** The broad user endpoint returns all of the user's repos regardless of
    the declared owners. **Resolution:** treat the declared owners as an allowlist applied after
    discovery (filter by owner login); the endpoint-selection logic still uses the list to decide
    which org endpoints to call.

29. **Orphan-confirmation endpoint.** One review flagged the id-based confirmation endpoint as
    undocumented; the other accepted it but wanted error handling. **Resolution:** keep the
    id-based endpoint (it is consistent with the id-stability invariant and avoids reintroducing
    the rename problem), record the undocumented risk explicitly, isolate it behind the discovery
    port, and provide an owner/name verification with redirect handling as a degraded fallback.
    Error taxonomy: only a not-found result means orphan; unauthorized/forbidden/legal-hold/server
    /network results do not.

30. **Permanent orphan re-verification.** A confirmed orphan whose clone stays on disk is
    re-verified on every run, with no way to silence it. **Resolution (negotiated — see trade-off
    3):** silence via a config-level ignore by id, which suppresses the whole pipeline for that id.
    Prune stays scoped to records whose path is already gone.

## 6. State / locking / testability

31. **Locking.** The plan was internally inconsistent (optional in one place, mandatory in
    another). **Resolution:** the file lock is mandatory and covers the whole run (not just the
    final flush), so two processes cannot plan the same side effects; failure to acquire fails
    fast with exit `2`.

32. **Golden-test determinism.** The JSON output embeds a generated id and timestamps, which
    defeats deterministic golden tests. **Resolution:** inject a clock and an id generator at the
    service/reporter layer (not globals); golden tests render a ready, deterministic plan so
    renderer tests are not mixed with plan-time id/clock generation.

## Negotiated trade-offs

The two reviews mostly complemented each other; three points were genuine trade-offs and were
resolved by converging on the narrower, safer option.

1. **Git transport auth breadth.** One side proposed a broad SSH surface (host-key modes,
   passphrase env, key management); the other argued for narrowing to HTTPS + PAT as the
   supported path. **Agreed:** the transport is selectable per profile; the supported path for
   private repos is HTTPS + PAT, which the private-repo success criterion relies on; SSH is
   best-effort via an existing agent and strict host-key checking, with the prerequisite stated
   that SSH is configured outside the tool. No key/passphrase management or host-key mutation in
   the PoC.

2. **Auto-moving external paths.** One side allowed the fetcher to move any path it had a record
   for; the other warned against ever relocating user directories outside the root. **Agreed:**
   external override paths are adopt/fetch-only; auto-move outside the root is deferred behind a
   future opt-in (not added now), and a changed override is surfaced as drift rather than executed.

3. **Acknowledging an orphan.** One side proposed a state "forget" command; the other showed it
   is insufficient because the on-disk clone is re-discovered next run. **Agreed:** silence an
   orphan via a config-level ignore by id (durable, version-controllable, visible), keeping prune
   for missing-path records only.

## Agreed specification changes

The non-trivial edits to the source plan:

1. §2.1 — drop the `afero`-for-GitSync claim and the "`slog` is the existing convention" framing;
   defer `slog` to the polish milestone.
2. §2.2 — make the state-file lock dependency mandatory (a new direct dependency).
3. §2.3 — ports are new; split them across `internal/service/github`, `internal/service/git`, and
   the state package; remove "already exists".
4. §3 / §3.2 / §3.3 — exit-code mapping in the first milestone; drop the dry-run flag; scope prune
   to missing paths; config-init on an existing file exits `2` (with `--force` to overwrite).
5. §4.1 — a token is always required; no config and no owner flag → exit `2`.
6. §4.2 — add an ignore field to a repo match; allow per-profile transport; filters gate only new
   clone/adopt.
7. §4.3 — source visibility from the API visibility field (fallback to the private boolean);
   include the internal value; state the strict path-resolution rule (owner-level stays within the
   root; absolute/`~` only for per-repo overrides).
8. §4.4 / §9 — pass 3: rename fallback via API redirect with ambiguity → conflict; index the disk
   scan by id including tracked repos; add a state-relocation action; external paths are
   adopt/fetch-only.
9. §5 — declared owners are an allowlist (post-discovery filter); add a Git transport auth
   subsection (HTTPS + PAT supported, SSH best-effort prerequisite); transport is per profile.
10. §6.2 — add the source-profile field; the stored remote URL is canonical and credential-free,
    with redaction at adoption.
11. §7 — add the collision-outcomes table; update-remote fires only on a canonical-URL change
    (transport from the observed value); the human plan collapses routine fetches while JSON lists
    them; the planner enforces a global target-path index before apply; add the relocation action.
12. §11 — bounded concurrency via the limit setter (and `concurrency >= 1`); a cap on
    rate-limit waiting when no wall-clock budget is set.
13. §14 — inject a clock and id generator for golden-test determinism; test GitSync via billy /
    temp directories rather than `afero`.
14. §16 — non-goals: the working tree is never touched; auto-move of external paths is deferred
    behind a future opt-in; an orphan is silenced via config-level ignore.

All 32 findings and the 3 trade-offs are closed; no open disagreements remain.
