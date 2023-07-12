---
id: 63
database_id: 1281851164
node_id: I_kwDOE2M9Zc5MZ38c
status: closed
title: "github: contribution: highlight suggested day"
labels: [scope: code, type: feature, impact: medium, effort: easy]
url: https://github.com/octomation/maintainer/issues/63
created_at: 2022-06-23T06:00:34Z
updated_at: 2023-04-05T20:27:04Z
---

# github: contribution: highlight suggested day

**Motivation:** check correctness visually.

**Interface**

```bash
$ maintainer github contribution suggest --delta 2013-11-20

 Day / Week    #45    #46    #47    #48   #49
------------- ------ ------ ------ ----- -----
 Sunday         -      -      ★      1     -
 Monday         -      -      -      2     1
 Tuesday        -      -      -      8     1
 Wednesday      -      1      1      -     3
 Thursday       -      -      3      7     1
 Friday         -      -      -      1     2
 Saturday       -      -      -      -     -
------------- ------ ------ ------ ----- -----
 Contributions for 2013-11-17: -3119d, 0 → 5
```
