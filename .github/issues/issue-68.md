---
id: 68
database_id: 1312861416
node_id: I_kwDOE2M9Zc5OQKzo
status: closed
title: "github: contribution: invalid suggestion for edge case with zero"
labels: [scope: code, scope: test]
url: https://github.com/octomation/maintainer/issues/68
created_at: 2022-07-21T08:15:21Z
updated_at: 2022-07-22T18:18:53Z
---

# github: contribution: invalid suggestion for edge case with zero

**Steps to reproduce**

```bash
$ maintainer github contribution suggest --delta 2021
 Day / Week   #35   #36   #37   #38   #39
------------ ----- ----- ----- ----- -----
 Sunday        6     7     9     6     6
 Monday        6     7     5     6     6
 Tuesday       6     7    12     6     5
 Wednesday     6     7    10     4     5
 Thursday      6     7     6     6     5
 Friday        6     7     4     5     6
 Saturday      6     -     6     1     -
------------ ----- ----- ----- ----- -----
 Suggestion is 2021-09-12: -312d, 9 → 12
```

Expected suggestion: `Suggestion is 2021-09-11: -313d, 0 → 7`
