---
code: MAIN-130
id: I_kwDOE2M9Zc8AAAABUzHaGA
databaseId: 5690743320
number: 468
url: https://github.com/octomation/maintainer/issues/468
title: "status: inspection: detect renames the same way in every count"
issueType: null
labels:
  - "scope: docs"
  - "scope: code"
  - "scope: test"
  - "type: bug"
  - "severity: minor"
  - "impact: medium"
  - "effort: easy"
assignees:
  - kamilsk
milestone: "[[MAIN-M3]]"
parent: null
issueFields: {}
state: CLOSED
stateReason: COMPLETED
duplicateOf: null
createdAt: 2026-10-03T14:58:12Z
updatedAt: 2026-10-19T07:36:32Z
lastEditedAt: 2026-10-19T07:36:32Z
closedAt: 2026-10-19T07:11:01Z
---

# status: inspection: detect renames the same way in every count

**Motivation:** one checkout must show the same numbers on every machine, and the numbers of one row must agree with each other. Today `f` depends on the user's `status.renames`, while `+/-` and `binary` always ignore renames, so a single `git mv` of a binary file reads `1f · 2 binary`, and a script that filters on `changed_files` gets different lists on different machines.

**Requirements**

- Both Git calls behind a row detect renames explicitly and identically, following Git's default: `git status` and the line count pass `--find-renames`; no user setting of `status.renames` or `diff.renames` changes a row.
- A staged rename is one changed file in `f`, its content changes only in `+/-`, one file in `binary`; a pure rename reads `+0/-0 · 1f`.
- The parsers accept the rename records of both commands under `-z`: `2 …\0<path>\0<orig>\0` of `git status` and `<a>\t<d>\t\0<orig>\0<path>\0` of `git diff --numstat`, including paths that contain a newline, a tab, or look like another record (`u old\nfile`).
- The status documentation describes the new counting, including the cases Git cannot detect: a rename not yet added to the index, a file rewritten beyond the similarity threshold.

**Evidence:** `internal/service/status/git.go:97` runs `git status` without a rename option, `git.go:157` runs the line count with `--no-renames`; `docs/content/status.md:59-61` documents rename detection as disabled. Found during the consultation on [[MAIN-127]], reproduced on `932b9c8`.

**PoC**

```bash
export GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_NOSYSTEM=1 XDG_STATE_HOME=$(mktemp -d)
tmp=$(mktemp -d) && d=$tmp/public/acme/tool && mkdir -p "$d" && cd "$d"
git init -q -b main && printf 'bin\0' > logo.png && seq 1 20 > app.go
git add -A && git -c user.name=T -c user.email=t@e.x commit -qm base
git remote add origin https://github.com/acme/tool.git
git mv logo.png icon.png && git mv app.go main.go
for v in true false; do
  GIT_CONFIG_COUNT=1 GIT_CONFIG_KEY_0=status.renames GIT_CONFIG_VALUE_0=$v \
    maintainer status --config="" --root "$tmp" --format=json | jq -c '.[] | {changed_files, added, deleted, binary}'
done
# today:    {"changed_files":2,"added":20,"deleted":20,"binary":2} then {"changed_files":4,…,"binary":2}
# expected: {"changed_files":2,"added":0,"deleted":0,"binary":1} both times
```

**Out of scope:** copy detection; pinning `status.renameLimit` and `diff.renameLimit`.

Related: [[MAIN-142]], [[MAIN-141]].
