---
id: 155
database_id: 1843914802
node_id: I_kwDOE2M9Zc5t5-gy
status: open
title: "github: contribution: simplify arguments format"
labels: ["scope: code","type: bug","severity: major","impact: medium","effort: medium"]
url: https://github.com/octomation/maintainer/issues/155
created_at: 2023-08-09T19:54:39Z
updated_at: 2023-08-09T19:54:40Z
---

# github: contribution: simplify arguments format

Fix the lookup crash on a short argument and make the period formats predictable. The original title speaks about simplifying the input, but the [screenshot](https://github.com/octomation/maintainer/assets/1165416/01fa06e5-7a9f-42e2-b18e-688956177309) records a concrete, reproducible bug:

```bash
maintainer github contribution lookup 2022
# recovered: assertion is not a true
# unexpected panic occurred
```

**Confirmed against the current code:** the call crashes before any GitHub data is fetched. With no suffix, the parser turns on the centred window while lookup passes a negative default number of weeks, and the range check rejects that combination. The same crash therefore also affects `lookup 2022-02`, `lookup 2022-02-01`, and a bare `lookup` with no argument at all — the very form the documentation advertises. Meanwhile the help explains no date arguments.

The expectation is no panic for a year, a month, a day, or empty input. The supported forms must have a clear meaning and examples in the help, and invalid input must produce an ordinary diagnostic error. The calendar-month contract is specified separately in [#45](issue-45.md); simply accepting `YYYY-MM` must not be mistaken for implementing it. The difference from [#121](issue-121.md): there the parsing of `2022/+10` was refused, whereas here the window calculation itself breaks when the suffix is absent.
