---
code: MAIN-153
id: null
databaseId: null
number: null
url: null
title: "fetch: config: reject a broken per-repository path template and report it per repository"
issueType: Bug
labels:
  - "type: bug"
  - "scope: code"
  - "scope: test"
  - "scope: docs"
  - "severity: minor"
  - "impact: medium"
  - "effort: easy"
assignees:
  - kamilsk
milestone: "[[MAIN-M3]]"
parent: null
issueFields: {}
state: OPEN
stateReason: null
duplicateOf: null
createdAt: null
updatedAt: null
lastEditedAt: null
closedAt: null
---

# fetch: config: reject a broken per-repository path template and report it per repository

## Problem

A per-repository `path` that is not a valid template, such as `~/ext/{{.Owner`, is handled in three different ways. `maintainer fetch config validate` accepts it, so the check meant to catch config mistakes says `ok`. `fetch` turns it into a conflict of that repository and plans the rest. `status` stops on it with exit `1` and shows no row at all, so one typo in one rule hides the whole inventory. `workspace.path` and the owner templates are already rejected by `validate`; the per-repository path is the only template it does not parse.

## PoC

```bash
set -o pipefail; tmp=$(mktemp -d) && mkdir -p "$tmp/ws" && export XDG_STATE_HOME="$tmp/state"
printf '[workspace]\nroot = "%s/ws"\n[[repos]]\nmatch = { owner = "acme", name = "broken" }\npath = "~/ext/{{.Owner"\n' "$tmp" > "$tmp/fetch.toml"
maintainer fetch --config "$tmp/fetch.toml" config validate; echo "exit=$?"
# today:    ok: …/fetch.toml (0 profiles, 0 owners, 1 repo rules), exit=0
# expected: an error naming repos[0], acme/broken and the parse error, exit=2
maintainer status --config "$tmp/fetch.toml" --format=plain; echo "exit=$?"
# today:    Error: parse path template "~/ext/{{.Owner": template: path:1: unclosed action, exit=1, no rows
# expected: every other row, plus an error row for acme/broken with the parse error, exit=1
```

## Expected behavior

- `maintainer fetch config validate` refuses a config whose per-repository `path` does not parse as a path template, with the functions path templates provide (`lower`, `upper`). The error names the rule (`repos[N]`), its match (`owner/name` or `id=N`) and the parse error; the exit code is `2`, as for every other config error.
- `fetch` and `status` don't refuse such a config: the broken rule is a problem of its repository only, and every other repository is handled as usual.
- `fetch` plans a conflict for that repository with the parse error as its reason, as for any path that cannot be rendered: the plan exits `0`, and `--apply` exits `3` for the unresolved conflict.
- `status` shows an error row for that repository with the rule's path as written and the parse error, inspects nothing in its place and renders every other row; like any other error row, it makes the command exit `1`. Unlike an `unresolved pin`, which waits for fetch to learn a field, a broken template is a mistake in the config that only the user can fix.

## Acceptance criteria

- [ ] The PoC's `validate` fails with the error above and exit `2`; a valid per-repository template, `lower` and `upper` included, still validates.
- [ ] A fetch planner test: a broken per-repository template yields one conflict carrying the parse error and a plan for the other repositories.
- [ ] A status test: a broken per-repository template, with and without a state record of the repository, yields one error row with the rule's path and the parse error next to the other rows, without an error from the collection.
- [ ] The fetch guide says that `fetch config validate` rejects a per-repository path that is not a valid template, and the status guide that `status` reports it as an error row.

## Out of scope

Unknown fields in a template that parses, such as `{{.Ownr}}`: they fail when the path is rendered, as before. The CUE schema of the config, which checks the structure, not the template syntax.

Related: [[MAIN-145]], [[MAIN-149]].
