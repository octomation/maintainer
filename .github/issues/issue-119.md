---
id: 119
database_id: 1633416767
node_id: I_kwDOE2M9Zc5hW_Y_
status: closed
title: "github: contribution: invalid suggestion for specific date"
labels: [type: bug, severity: critical, impact: high, effort: medium]
url: https://github.com/octomation/maintainer/issues/119
created_at: 2023-03-21T08:18:46Z
updated_at: 2023-03-25T20:26:23Z
---

# github: contribution: invalid suggestion for specific date

**Bad case**

```bash
$ maintainer github contribution suggest --delta 2022-02-12
 Day / Week   #6   #7   #8   #9   #10   #11
------------ ---- ---- ---- ---- ----- -----
 Sunday       6    6    6    6     4     5
 Monday       6    6    5    6     5     5
 Tuesday      6    6    6    6     7     5
 Wednesday    1    5    6    3     5     5
 Thursday     1    6    6    6     5     5
 Friday       5    5    6    6     5     5
 Saturday     10   6    6    5     5     5
------------ ---- ---- ---- ---- ----- -----
 Suggestion is 2022-02-06: -408d, 6 → 10
```

Must be 2022-02-16 with 5 → 6, not 2022-02-06.

Related to #84 and kamilsk/dotfiles/issues/543.
