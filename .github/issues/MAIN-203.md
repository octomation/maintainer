---
code: MAIN-203
id: I_kwDOE2M9Zc8AAAABjEVxEA
databaseId: 6648328464
number: 581
url: https://github.com/octomation/maintainer/issues/581
title: "skills: review: read the GitHub security alerts in security mode"
issueType: Task
labels:
  - "effort: medium"
  - "impact: medium"
  - "scope: docs"
  - "scope: test"
assignees:
  - kamilsk
milestone: null
parent: null
issueFields: {}
state: OPEN
stateReason: null
duplicateOf: null
createdAt: 2026-11-27T14:24:49Z
updatedAt: 2026-11-27T18:41:43Z
lastEditedAt: 2026-11-27T18:41:43Z
closedAt: null
---

# skills: review: read the GitHub security alerts in security mode

## Motivation

GitHub already watches the dependencies of a repository: Dependabot alerts,
code scanning and secret scanning report known vulnerabilities and leaked
secrets. A security review of the skill runs its own audits and never looks
at them, so the owner checks the GitHub notifications by hand, and a dismissed
alert, such as the zod one of the docs site, has no place in the report.

## Scope

- Security mode reads the alerts of the repository through `gh`, read-only:
  open and dismissed Dependabot alerts, open code scanning and secret scanning
  alerts, and draft repository security advisories.
- An open alert is a candidate of the Supply chain lens: the reviewer verifies
  whether the product ships the vulnerable code, as with any candidate.
- A dismissed alert with a reason and a comment that still apply is an owner's
  decision: it gives `Accepted`, not a finding. A dismissal without a reason,
  or with a reason that no longer applies, is a candidate.
- The Gates lens names the counts by state and the scanners that are off.
- Without `gh`, without access or without a GitHub remote, the gate is
  `partial`.

## Acceptance criteria

- [x] `review.py alerts` prints one line per open or dismissed alert and one
      summary line per source, uses only `gh api` reads, and exits non-zero
      when it cannot read a source that is on.
- [x] A disabled scanner prints as `off`, not as an error.
- [x] `references/security.md` says how an alert becomes a candidate, an
      `Accepted` decision or nothing, and how the Gates lens reports them.
- [x] The tests of `review.py` cover `alerts` with a fake `gh`.
- [ ] A security review of this repository names the dismissed zod alert in
      the evidence of the Gates lens.

## Out of scope

Changing the state of an alert. Notifications of other kinds, such as
mentions or CI failures.
