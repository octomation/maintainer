---
id: 76
database_id: 1322302350
node_id: I_kwDOE2M9Zc5O0LuO
status: closed
title: "github: contribution: support strict ISO 8601 format as input date"
labels: [scope: code, type: feature, impact: medium, effort: easy]
url: https://github.com/octomation/maintainer/issues/76
created_at: 2022-07-29T14:10:53Z
updated_at: 2023-03-26T20:22:55Z
---

# github: contribution: support strict ISO 8601 format as input date

**Motivation:** suggestion is not accurate. It's based on current time, but it's not optimal

1. use + ~(working hours) for "new" date
2. use + ~delta for latest commit

**Interface**

`maintainer github contribution suggest --short "$(git --no-pager log -1 --format="%aI")`
