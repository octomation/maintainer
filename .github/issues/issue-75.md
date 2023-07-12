---
id: 75
database_id: 1318589500
node_id: I_kwDOE2M9Zc5OmBQ8
status: closed
title: "github: contribution: suggest shows future"
labels: [scope: code, type: bug, severity: minor, impact: low, effort: easy]
url: https://github.com/octomation/maintainer/issues/75
created_at: 2022-07-26T18:07:42Z
updated_at: 2023-04-05T20:54:21Z
---

# github: contribution: suggest shows future

**Steps to reproduce**

```bash
$ maintainer github contribution suggest --delta --target=10 2022-05-01/+12
Day / Week   #18   #19   #20   #21   #22   #23   #24   #25   #26   #27   #28   #29   #30
------------ ----- ----- ----- ----- ----- ----- ----- ----- ----- ----- ----- ----- -----
 Sunday       10     6     6     5     5     5     5     5     5     3     5     4     5
 Monday       10     4     3     5     4     5     3     5     5     6     5     5    10
 Tuesday      10     6     6     5     5     4     5     5     4     6     5     5     1
 Wednesday     4     6     6     5     5     5     5     5     4     6     5     5     -
 Thursday      6     6     6     5     4     4     3     4     4     6     -     5     -
 Friday        6     6     6     5     5     5     4     5     4     6     1     4     -
 Saturday      6     6     6     5     5     4     3     5     2     6     -     5     -
------------ ----- ----- ----- ----- ----- ----- ----- ----- ----- ----- ----- ----- -----

$ maintainer github contribution suggest --delta --target=10 2022-07-24
 Day / Week   #30   #31   #32   #33   #34   #35
------------ ----- ----- ----- ----- ----- -----
 Sunday        5     -     -     -     -     -
 Monday       10     -     -     -     -     -
 Tuesday       1     -     -     -     -     -
 Wednesday     -     -     -     -     -     -
 Thursday      -     -     -     -     -     -
 Friday        -     -     -     -     -     -
 Saturday      -     -     -     -     -     -
------------ ----- ----- ----- ----- ----- -----
```

**Expected**

- For 2022-07-27 `-` -> `?`
- `#31` and rest must be hidden
