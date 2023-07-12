---
id: 124
database_id: 1652573543
node_id: I_kwDOE2M9Zc5igEVn
status: closed
title: "github: contribution: suggest use incorrect center"
labels: [type: bug, severity: major, impact: medium, effort: medium]
url: https://github.com/octomation/maintainer/issues/124
created_at: 2023-04-03T18:36:14Z
updated_at: 2023-04-05T19:40:33Z
---

# github: contribution: suggest use incorrect center

**Details**

```bash
$ git at $(suggest) 'tools(deps): bump github.com/evanw/esbuild from 0.17.14 to 0.17.15'
 Day / Week   #32   #33   #34   #35   #36   #37   #38   #39   #40   #41   #42
------------ ----- ----- ----- ----- ----- ----- ----- ----- ----- ----- -----
 Sunday        5     5     5     5     5     5     3     5     5     5    15
 Monday        5     5     5     5     5     5     5     5     5     5     7
 Tuesday       5     5     5     5     5     5     5     5     5     5     4
 Wednesday     5     5     5     5     5     5     5     5     5     5    10
 Thursday      5     5     5     5     5     5     5     5     5     5    11
 Friday        5     5     5     5     5     5     5     5     5     5    11
 Saturday      5     5     5     5     5     6     5     5     5     5    11
------------ ----- ----- ----- ----- ----- ----- ----- ----- ----- ----- -----
 Contributions are on the range from 2022-08-07 to 2022-10-22

$ git at $(suggest) 'tools(deps): bump github.com/mikefarah/yq/v4 from 4.33.1 to 4.33.2'
 Day / Week   #32   #33   #34   #35   #36   #37   #38   #39   #40   #41   #42
------------ ----- ----- ----- ----- ----- ----- ----- ----- ----- ----- -----
 Sunday        5     5     5     5     5     5     4     5     5     5    15
 Monday        5     5     5     5     5     5     5     5     5     5     7
 Tuesday       5     5     5     5     5     5     5     5     5     5     4
 Wednesday     5     5     5     5     5     5     5     5     5     5    10
 Thursday      5     5     5     5     5     5     5     5     5     5    11
 Friday        5     5     5     5     5     5     5     5     5     5    11
 Saturday      5     5     5     5     5     6     5     5     5     5    11
------------ ----- ----- ----- ----- ----- ----- ----- ----- ----- ----- -----
 Contributions are on the range from 2022-08-07 to 2022-10-22
```

the center is `#38` week, not `#37`.
