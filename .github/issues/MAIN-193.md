---
code: MAIN-193
id: I_kwDOE2M9Zc8AAAABivK46Q
databaseId: 6626130153
number: 566
url: https://github.com/octomation/maintainer/issues/566
title: "skills: maintainer: cover the configuration keys, exit codes and verbose output that users need"
issueType: Feature
labels:
  - "scope: docs"
  - "scope: test"
  - "type: feature"
  - "impact: medium"
  - "effort: medium"
assignees:
  - kamilsk
milestone: null
parent: null
issueFields: {}
state: CLOSED
stateReason: COMPLETED
duplicateOf: null
createdAt: 2026-11-26T11:56:43Z
updatedAt: 2026-11-26T15:14:44Z
lastEditedAt: 2026-11-26T15:14:43Z
closedAt: 2026-11-26T15:02:01Z
---

# skills: maintainer: cover the configuration keys, exit codes and verbose output that users need

## Motivation

`go test ./skills` keeps the command reference of the agent skill in step
with the CLI, but not the meaning of keys, fields and exit codes. An agent
with the skill cannot help with several common cases:

- A confirmed `orphan` or a checkout behind SSO: the skill does not know
  `[[repos]] ignore = true`.
- A private repository that fails to clone: the skill does not know
  `clone_url` (ssh or https) or `[[owners]] path`. It points to `fetch.cue`,
  which a user of the binary does not have.
- The service under `brew services` does not start: the skill does not know
  `[fetch] config` and `[[repos]] weight` in `service.toml`, or that the
  service needs a fetch config.
- A failure to diagnose: §7 does not name `-vv` and `-vvv`.
- Output to parse: `references/output.md` lacks the exit codes of `skill` and
  `service` and the `id` field of `status --format=json`.

## PoC

```bash
maintainer fetch              # a confirmed orphan in every plan
# the user asks the agent to stop it showing up; today the skill has no answer
```

## Acceptance criteria

- [x] The skill names `ignore = true`, `clone_url` and `[[owners]] path` with
      one line each, and links the configuration section of the fetch guide
      instead of `fetch.cue`.
- [x] The service section names `[fetch] config` and `[[repos]] weight`, and
      what the service does without a fetch config.
- [x] §7 names `-vv` and `-vvv`, and `service open --listen` for a service on
      another port.
- [x] `references/output.md` has the exit codes of `skill` and `service` and
      every field of `status --format=json`.
- [x] A test fails when a JSON field that `references/output.md` names is
      missing from the output structures. If this test is not small, it moves
      to its own issue and this criterion is dropped.
- [x] `go test ./skills` passes.

## Out of scope

How the agent works with the service beyond these keys: that is decided
separately.

Related: [[MAIN-191]], [[MAIN-192]].
