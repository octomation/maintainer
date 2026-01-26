---
code: MAIN-146
id: null
databaseId: null
number: null
url: null
title: "fetch: state: write the state file without following a planted link"
issueType: Bug
labels:
  - "type: bug"
  - "scope: code"
  - "scope: test"
  - "severity: minor"
  - "impact: low"
  - "effort: easy"
assignees:
  - kamilsk
milestone: null
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

# fetch: state: write the state file without following a planted link

## Problem

Saving the state must never touch a file outside it, but a symlink planted next to the state file, at `<state_file>.tmp`, is followed: its target is overwritten with the state JSON and set to mode 0600, and the state file itself becomes that link. Every save does it, `state prune` included. The default state directory is private, so planting the link takes the same user or a `state_file` in a shared directory — an exotic setup, but the overwritten file cannot be restored.

## PoC

```bash
set -o pipefail; tmp=$(mktemp -d) && export XDG_STATE_HOME="$tmp/xdg"
mkdir -p "$tmp/st" && printf 'precious\n' > "$tmp/victim.txt"
printf '[defaults]\nstate_file = "%s/st/state.json"\n[profiles.primary]\ntoken_env = "GITHUB_TOKEN"\n' "$tmp" > "$tmp/fetch.toml"
printf '{"version":1,"profiles":{},"repos":[]}\n' > "$tmp/st/state.json" && chmod 600 "$tmp/st/state.json"
ln -s "$tmp/victim.txt" "$tmp/st/state.json.tmp"
maintainer fetch --config "$tmp/fetch.toml" state prune
cat "$tmp/victim.txt"; ls -l "$tmp/st"
# today:    victim.txt holds the state JSON with mode 0600, and state.json is a symlink to it
# expected: victim.txt still says "precious", and state.json is a regular 0600 file
```

## Expected behavior

A save never writes through a path that existed before it started: whatever lies at the old temporary name — a file or a link — is neither followed, truncated nor chmodded. The result is a regular file with mode 0600 at the state path, and a save that fails leaves the previous state file as it was.

## Acceptance criteria

- [ ] The PoC leaves `victim.txt` unchanged, and `state.json` is a regular file with mode 0600.
- [ ] A store test plants a symlink and, separately, a regular file at `<state_file>.tmp`; `Save` leaves both targets byte-identical and never leaves the state path as a link.
- [ ] A store test with a save that fails midway keeps the previous state file byte-identical.

## Out of scope

The location, the format and the locking of the state file; the permissions of a directory the user chose for it. The other state safety work is [[MAIN-137]].

## Sources

- [[20261019T105848Z#F1. The state temporary file follows a symlink]]
