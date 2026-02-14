---
code: MAIN-171
id: null
databaseId: null
number: null
url: null
title: "skill: status: report unknown compatibility for a pseudo-version binary"
issueType: Bug
labels:
  - "scope: code"
  - "scope: test"
  - "type: bug"
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

# skill: status: report unknown compatibility for a pseudo-version binary

## Problem

A binary built from an untagged commit reports a pseudo-version, such as
`v0.5.1-0.20261125120000-1848546abcde`. `skill status` then calls the bundled
skill of that binary incompatible, exits `1` and advises `install --replace`,
which changes nothing. `skills/maintainer/SKILL.md` tells the agent that a
pseudo-version means unknown compatibility.

## PoC

```bash
tmp=$(mktemp -d)
go build -ldflags "-X main.version=v0.5.1-0.20261125120000-1848546abcde -X main.commit=1848546" -o "$tmp/maintainer" .
"$tmp/maintainer" skill status; echo "exit=$?"
# today:    "incompatible with this binary, same as this release", exit=1
# expected: compatibility unknown, exit=0
```

## Expected behavior

For a binary whose version is `dev`, a pseudo-version or unknown, `skill status`
reports the compatibility of each copy as unknown and does not fail on it. The
other checks keep their result: a changed copy, a missing copy for `--agent`, and
`--exact`.

## Acceptance criteria

- [ ] With a copy installed from the PoC binary, the PoC reports unknown
      compatibility and exits `0`.
- [ ] A test covers a pseudo-version, `dev` and a prerelease tag; the prerelease
      tag keeps the range check.
- [ ] `skill status` and `skills/maintainer/SKILL.md` give a pseudo-version the
      same meaning.

Source: [20261125T152232Z](../reviews/20261125T152232Z.md), "skill status calls the skill of a pseudo-version binary incompatible".
