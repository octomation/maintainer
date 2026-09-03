---
code: MAIN-123
id: 445
database_id: 5142953708
node_id: I_kwDOE2M9Zc8AAAABMos-7A
status: open
title: "fetch: prevent destructive clone reconciliation for occupied targets"
labels: ["scope: code","scope: test","type: bug","severity: critical","impact: high","effort: medium"]
url: https://github.com/octomation/maintainer/issues/445
created_at: 2026-08-13T15:21:42Z
updated_at: 2026-08-13T15:28:43Z
---

# fetch: prevent destructive clone reconciliation for occupied targets

**Motivation:** `fetch --apply` must uphold its non-destructive contract even when checkout discovery is ambiguous or a reviewed plan becomes stale. A push-only remote URL used to prevent accidental pushes must not affect checkout identity.

**Details**

Git fetch URLs and push URLs are independent configuration. If inspection conflates them, an existing checkout can be treated as missing and a clone can be planned at the same path. If clone-error cleanup then recursively removes that path, a read-only reconciliation mistake becomes local data loss.

Required invariants:

- repository identity uses fetch URLs only; push URLs remain orthogonal and survive remote updates;
- plan output reports an explicit conflict, including the path, whenever a clone or move target exists but cannot be verified as the expected repository;
- apply revalidates the target and refuses to overwrite anything that appeared after planning;
- failure cleanup never recursively removes a pre-existing or unverifiable path;
- existing push-lock workflows continue to permit fetch while preventing push.

**PoC**

```bash
go test ./internal/service/git ./internal/service/fetch \
  -run 'PushURL|PushLock|ExistingTarget|OccupiedCloneTarget' \
  -count=1
```

The regression suite must demonstrate that locked checkouts reconcile normally, occupied targets are visible during planning, and local-only files survive a refused clone.
