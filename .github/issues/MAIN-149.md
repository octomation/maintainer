---
code: MAIN-149
id: I_kwDOE2M9Zc8AAAABYusKIg
databaseId: 5954538018
number: 505
url: https://github.com/octomation/maintainer/issues/505
title: "fetch: config: reject unknown keys and keep a CUE schema next to the config"
issueType: Bug
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
milestone: "[[MAIN-M2]]"
parent: null
issueFields: {}
state: CLOSED
stateReason: COMPLETED
duplicateOf: null
createdAt: 2026-10-21T04:49:59Z
updatedAt: 2026-10-21T10:23:41Z
lastEditedAt: null
closedAt: 2026-10-21T04:50:02Z
---

# fetch: config: reject unknown keys and keep a CUE schema next to the config

## Problem

Unknown keys in the fetch config are ignored without a word, so a typo goes unnoticed. A profile with `include_owner = ["acme"]` validates, and since an omitted `include_owners` means "you and every organisation you are a member of", the typo widens discovery instead of narrowing it. A key outside the schema never had an effect, so no working configuration relies on it: rejecting it turns the implicit contract into an explicit one. The contract itself is a CUE schema kept next to the config, as `.github/settings.cue` is for `.github/settings.json`.

## PoC

```bash
set -o pipefail; tmp=$(mktemp -d) && export XDG_STATE_HOME="$tmp/xdg"
printf '[profiles.primary]\ntoken_env = "GITHUB_TOKEN"\ninclude_owner = ["acme"]\n' > "$tmp/fetch.toml"
maintainer fetch --config "$tmp/fetch.toml" config validate; echo "exit=$?"
# today:    ok: …/fetch.toml (1 profiles, 0 owners, 0 repo rules), exit=0, nothing about include_owner
# expected: an error naming include_owner, [profiles.primary] and the file, suggesting include_owners; exit=2
```

## Expected behavior

Every command that loads the fetch config — `fetch` and its subcommands, and `status` — refuses a config with an unknown key or table as bad config (exit `2`), naming each one: the key, its table, the file and, when one is close, the known key it was probably meant to be. The keys, their types and defaults are defined by a CUE schema in the repository; the loader, the template and the guides agree with it.

## Acceptance criteria

- [ ] The PoC fails with the error above and exit `2`.
- [ ] The same for a YAML config, for an unknown table and for unknown keys inside `[filters]` and `[[repos]]`.
- [ ] A CUE schema of the fetch config lives in the repository; the template written by `maintainer fetch config init` and every config example of the guides validate against it in a test, and load without an error.
- [ ] `fetch.md` says that unknown keys are rejected and where the schema is.

## Out of scope

A config migration; editor integration of the schema. Owner selection itself is [[MAIN-133]].

## Sources

- [[20261019T105848Z#F5. Unknown configuration keys are accepted silently]]
