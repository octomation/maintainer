---
code: MAIN-145
id: null
databaseId: null
number: null
url: null
title: "workspace: paths: resolve pins only with sufficient context"
labels:
  - "type: bug"
  - "scope: code"
  - "scope: test"
  - "severity: major"
  - "impact: medium"
  - "effort: easy"
milestone: "[[MAIN-M3]]"
state: OPEN
stateReason: null
createdAt: 2026-09-26T09:27:00Z
updatedAt: 2026-09-26T09:27:00Z
lastEditedAt: null
closedAt: null
---

# workspace: paths: resolve pins only with sufficient context

**Motivation:** a per-repository `path` is documented as a template rendered with the full context, and a pin must either select an existing checkout or say that it cannot. A pin rendered from a truncated snapshot, a path collapsed from empty template fields, one unrenderable path that ends the whole plan, and a path outside the root shortened by string prefix all contradict that.

**Requirements**

- Precedence of a pin: the currently applicable rule, rendered with sufficient context, comes first; the persisted `pinned_path` is the fallback only when no applicable rule replaces it. Changing a rule therefore selects another existing verified checkout, as the fetch specification §17 permits, and the previous path stays in the diagnostics.
- Sufficient context means every referenced field is known from the configuration (`.Root`), from the match rule (owner and name of a name-based rule) or from the state record (visibility, default branch, flags); a record is not required when the rule alone resolves the fields. Fetch completes a snapshot from the record before rendering and never renders from `{id, owner, name}` alone.
- Offline `status` uses a rule's path as a pin only under the same condition; otherwise it shows an "unresolved pin" row and inspects nothing.
- A per-snapshot `Resolve` error becomes a conflict action carrying the reason for that repository only.
- Paths in human output are shortened only when `workspace.Within(root, path)` holds, by path component.
- Tests: a templated pin applied twice exits `0` on the second run; changing an existing pin switches to the other checkout and keeps the previous path in diagnostics; a name-based rule without state resolves; an id-only templated rule without state is unresolved; a sibling path outside the root is printed in full.

**Evidence:** findings 7 (`internal/service/fetch/service.go:141`, `pinned.go:15`), 16 (`internal/service/status/service.go:99`), 22 (`internal/service/fetch/planner.go:278`) and 24 (`internal/service/fetch/reporter.go:244`) of [the code review](../reviews/20260926T071417Z.md), reproduced on `932b9c8`.

**PoC**

```bash
set -o pipefail; tmp=$(mktemp -d) && mkdir -p "$tmp/ws" && export XDG_STATE_HOME="$tmp/state" GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_NOSYSTEM=1
printf '[workspace]\nroot = "%s/ws"\n[[repos]]\nmatch = { id = 42 }\npath = "%s/ext/{{.Owner}}/{{.Repo}}"\n' "$tmp" "$tmp" > "$tmp/fetch.toml"
maintainer status --config "$tmp/fetch.toml" --format=plain
# today: ext │ — │ ? │ error: stat …/ext/.git: no such file or directory, exit 1
# expected: one "unresolved pin" row without inspection, exit 0
```

The phantom-conflict and single-render cases are reconstructed as regression tests of the fetch service and the planner from the review's descriptions (its scratch tests are not in the tree): a templated per-repo pin adopted on the first apply must plan only `fetch` on the second; a layout `{{.Owner}}/{{.Repo}}@{{.DefaultBranch}}` with a default branch `release/v1` must yield one conflict and a plan for the other repositories.

**Out of scope:** new template functions; moving pinned checkouts; multiple workspaces.

Related: [[MAIN-135]], [[MAIN-144]], [[MAIN-141]].
