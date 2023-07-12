---
id: 122
database_id: 1641614380
node_id: I_kwDOE2M9Zc5h2Qws
status: closed
title: "github: contribution: suggest works incorrectly after refactoring"
labels: [scope: code, scope: test, type: bug, severity: major, impact: medium, effort: easy]
url: https://github.com/octomation/maintainer/issues/122
created_at: 2023-03-27T08:11:43Z
updated_at: 2023-04-05T19:42:34Z
---

# github: contribution: suggest works incorrectly after refactoring

**Details**

```bash
$ maintainer github contribution suggest
 Day / Week      #28     #29    #30    #31   #32
-------------- ------- ------- ------ ----- -----
 Sunday           -       5      10     8     5
 Monday           5       5      10    10     5
 Tuesday          5       5      8     10     5
 Wednesday        5       5      10    10     5
 Thursday         5       5      6      9     5
 Friday           5       5      9     10     5
 Saturday         5       5      10    10     5
-------------- ------- ------- ------ ----- -----
 Suggestion is 2022-07-26T09:39:14+03:00, 0 → 10

$ git --no-pager log -1
commit b6efbfce77c6241e29c42685917be2308dd26b1e (HEAD -> main, tag: v0.1.0-rc10, origin/main)
Author: Kamil Samigullin <kamil@samigullin.info>
Date:   Tue Jul 26 08:57:56 2022 +0300

    fix #58: add jitter to suggested commiter date
```

Affected release: https://github.com/octomation/maintainer/releases/tag/v0.1.0-rc10.

---

I need two fixes:
- [x] `Sunday           -`
- [x] `0 → 10`
