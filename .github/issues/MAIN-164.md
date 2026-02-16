---
code: MAIN-164
id: I_kwDOE2M9Zc8AAAABihzV3A
databaseId: 6612112860
number: 550
url: https://github.com/octomation/maintainer/issues/550
title: "fetch: config: write a template that loads in the format of the config path"
issueType: Bug
labels:
  - "effort: easy"
  - "impact: medium"
  - "scope: code"
  - "scope: docs"
  - "scope: test"
  - "severity: minor"
  - "type: bug"
assignees:
  - kamilsk
milestone: null
parent: null
issueFields: {}
state: OPEN
stateReason: null
duplicateOf: null
createdAt: 2026-11-25T19:30:39Z
updatedAt: 2026-11-25T19:30:39Z
lastEditedAt: null
closedAt: null
---

# fetch: config: write a template that loads in the format of the config path

## Problem

`fetch config init` writes the TOML template to any path, a `.yaml` path
included, and reports success. Every later command then refuses that config as
bad config.

## PoC

```bash
tmp=$(mktemp -d)
maintainer fetch --config "$tmp/fetch.yaml" config init; echo "exit=$?"
maintainer fetch --config "$tmp/fetch.yaml" config validate; echo "exit=$?"
# today:    "wrote …/fetch.yaml", exit=0; then "parse YAML fetch config", exit=2
# expected: init leaves no file that validate refuses
```

## Expected behavior

`fetch config init` writes only a config that `fetch config validate` accepts at
the same path. For a path whose extension names a format without a template, it
writes the template in that format, or it refuses the path as a usage error
(exit `2`) and writes nothing.

## Acceptance criteria

- [x] In the PoC, `config validate` accepts the file that `init` wrote, or `init`
      exits `2`, names the extension and leaves no file.
- [x] A test runs `config init` for `.toml`, `.yaml` and `.yml` paths and loads
      each result, or expects the refusal.
- [x] `docs/content/fetch.md` names the formats that `config init` writes.

## Out of scope

The content of the template. The refusal to overwrite an existing file without
`--force`.

Source: [20261125T152232Z](../reviews/20261125T152232Z.md), "`fetch config init` writes the TOML template into a `.yaml` path".
