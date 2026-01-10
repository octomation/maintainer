---
code:
id: I_kwDOE2M9Zc5IgNqE
databaseId: 1216404100
number: 24
url: https://github.com/octomation/maintainer/issues/24
title: "use contributions chart for git at suggestion"
labels:
  - "scope: code"
milestone: "[[milestone-1]]"
state: CLOSED
stateReason: COMPLETED
createdAt: 2022-04-26T19:47:24Z
updatedAt: 2022-06-15T10:15:47Z
lastEditedAt: 2022-05-10T07:22:06Z
closedAt: 2022-06-02T15:23:55Z
---

# use contributions chart for git at suggestion

Find the gaps in the GitHub contributions calendar and propose a date to contribute, so the maintainer does not compare days by hand. The result must explain the choice: the date, the current count and the target value.

The working interface:

```bash
maintainer github contribution suggest --target=5 2021/+10
# Meaning of the result: a day with 3 contributions was chosen, the target is 5.
```

For automation, `--short` prints the date on stdout and `--delta` switches it to a relative representation. The calendar output makes neighbouring days reviewable without a browser. Passing the date into `git at` belongs to external Git wrappers.

**Current state:** the task is closed and the suggestion works, but open defects around boundaries and time are tracked separately, for example [#133](issue-133.md) and [#148](issue-148.md). Closure must not be read as a guarantee that every case is correct.

Original references: [github-contributions-chart](https://github.com/sallar/github-contributions-chart) and [github-calendar](https://github.com/Bloggify/github-calendar).
