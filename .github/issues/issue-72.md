---
id: 72
database_id: 1317178999
node_id: I_kwDOE2M9Zc5Ogo53
status: closed
title: "github: contribution: bad suggestion with Sunday shift"
labels: [scope: code, scope: test]
url: https://github.com/octomation/maintainer/issues/72
created_at: 2022-07-25T18:01:46Z
updated_at: 2022-07-25T18:33:26Z
---

# github: contribution: bad suggestion with Sunday shift

**Steps to reproduce**

```bash
$ maintainer github contribution suggest --delta 2022-05-01/+1
 Day / Week          #17        #18
----------------- ---------- ----------
 Sunday               3          6
 Monday               4          6
 Tuesday              5          6
 Wednesday            6          1
 Thursday             6          6
 Friday               6          6
 Saturday             4          6
----------------- ---------- ----------
 Suggestion is 2022-04-24: -92d, 3 → 6
```

Expected: `2022-05-04`, see

```bash
$ maintainer github contribution suggest --delta 2022-05-02/+1
 Day / Week          #18        #19
----------------- ---------- ----------
 Sunday               6          6
 Monday               6          4
 Tuesday              6          6
 Wednesday            1          6
 Thursday             6          6
 Friday               6          6
 Saturday             6          6
----------------- ---------- ----------
 Suggestion is 2022-05-04: -82d, 1 → 6
```
