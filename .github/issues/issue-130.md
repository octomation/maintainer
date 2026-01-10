---
code:
id: I_kwDOE2M9Zc5ixwje
databaseId: 1657211102
number: 130
url: https://github.com/octomation/maintainer/issues/130
title: "github: contribution: refactor histogram command"
labels:
  - "scope: code"
  - "scope: test"
  - "type: improvement"
  - "impact: medium"
  - "effort: easy"
milestone: "[[milestone-1]]"
state: OPEN
stateReason:
createdAt: 2023-04-06T11:06:57Z
updatedAt: 2023-04-06T11:06:58Z
lastEditedAt:
closedAt:
---

# github: contribution: refactor histogram command

Make the yearly histogram readable in an ordinary terminal. Every day is drawn as its own `#` today, so a frequent count produces a line hundreds of characters wide and destroys the overview of the distribution.

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

**At present** the output still prints one character per day. This is a presentation improvement for an existing command ([#29](issue-29.md)), not a change to how contributions are counted.
