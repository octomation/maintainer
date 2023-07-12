---
id: 126
database_id: 1656192148
node_id: I_kwDOE2M9Zc5it3yU
status: closed
title: "github: contribution: incorrect shows stats for future"
labels: [help wanted, scope: code, scope: test, type: bug, severity: minor, impact: low, effort: easy]
url: https://github.com/octomation/maintainer/issues/126
created_at: 2023-04-05T19:49:00Z
updated_at: 2023-04-05T20:02:21Z
---

# github: contribution: incorrect shows stats for future

**Details**

```bash
$ maintainer github contribution lookup /-2
 Day / Week   #12   #13   #14
------------ ----- ----- -----
 Sunday       10    10    10
 Monday       10     9    10
 Tuesday      10    10    10
 Wednesday    10    10     8
 Thursday     10    10     - # <- ?
 Friday       10    10     - # <- ?
 Saturday     10    10     - # <- ?
------------ ----- ----- -----
