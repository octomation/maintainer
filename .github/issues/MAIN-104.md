---
code: MAIN-104
id: I_kwDOE2M9Zc5ixwje
databaseId: 1657211102
number: 130
url: https://github.com/octomation/maintainer/issues/130
title: "github: contribution: refactor histogram command"
issueType: null
labels:
  - "scope: code"
  - "scope: test"
  - "type: improvement"
  - "impact: medium"
  - "effort: easy"
assignees:
  - kamilsk
milestone: "[[MAIN-M1]]"
parent: null
issueFields: {}
state: OPEN
stateReason: null
duplicateOf: null
createdAt: 2023-04-06T11:06:57Z
updatedAt: 2026-11-27T14:37:55Z
lastEditedAt: 2026-11-27T14:37:55Z
closedAt: null
---

# github: contribution: refactor histogram command

Make the yearly histogram readable in an ordinary terminal. Every day is drawn as its own `#`, so a frequent count produces a line hundreds of characters wide and destroys the overview of the distribution.

The original scenario:

```bash
maintainer github contribution histogram 2022
# The row for 5 contributions carried well over a hundred `#`, which is where the problem shows.
```

A prototype of compact output:

```text
Contributions  Days  Distribution
5              152   ####################
6               76   ##########
```

The numbers are illustrative. The bar length is scaled while the exact frequency stays a number, and a legend or a header explains the units. The order of values and `--with-zero` have to be preserved, and the behaviour in a narrow terminal and when redirected to a file has to be defined.

This is a presentation improvement for an existing command ([[MAIN-29]]), not a change to how contributions are counted.

## Acceptance criteria

- [x] `histogram` prints a header `Contributions  Days  Distribution`, the
      exact number of days in each row, and a bar scaled to at most 40
      characters, with at least one `#` for a row with days.
- [x] The order of rows and `--with-zero` stay as they are; empty data,
      zero-only data and a skewed distribution have tests.
- [x] The output is the same in a terminal and in a file, and goes through
      the writer of the command.
- [x] `docs/content/contributions.md` shows an example of the output, and
      `go test ./skills` passes with the generated reference.
