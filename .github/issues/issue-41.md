---
id: 41
database_id: 1243888256
node_id: I_kwDOE2M9Zc5KJDqA
status: closed
title: "github: contribution: lookup forward/backward"
labels: [scope: code]
url: https://github.com/octomation/maintainer/issues/41
created_at: 2022-05-21T06:31:57Z
updated_at: 2022-06-15T10:15:50Z
---

# github: contribution: lookup forward/backward

now it split by half, I need a possibility to look forward or backward

```bash
$ maintainer github contribution lookup 2021-01-01/10
 Day / Week       #53      #1      #2      #3      #4     #5
--------------- -------- ------- ------- ------- ------- -----
 Sunday            ?        5      10       5       5      5
 Monday            ?        5      10       5       5      5
 Tuesday           ?        5      10       5       5      5
 Wednesday         ?        5      10       5       5      5
 Thursday          ?        5      10       5       5      5
 Friday            5        5      10       5       5      5
 Saturday          5        5      10       5       5      5
--------------- -------- ------- ------- ------- ------- -----
 Contributions are on the range from 2021-01-01 to 2021-02-06
```

```bash
$ maintainer github contribution lookup 2021-01-01/10 forward
$ maintainer github contribution lookup /10 backward
```

or

```bash
$ maintainer github contribution lookup 2021-01-01/+10
$ maintainer github contribution lookup /-10
```

and possibly allow autodetection
