---
id: 73
database_id: 1317194684
node_id: I_kwDOE2M9Zc5Ogsu8
status: closed
title: "github: contribution: suggest command has regression with weeks arg"
labels: [scope: code]
url: https://github.com/octomation/maintainer/issues/73
created_at: 2022-07-25T18:16:08Z
updated_at: 2022-07-25T18:20:49Z
---

# github: contribution: suggest command has regression with weeks arg

**Steps to reproduce**

```bash
$ maintainer version
maintainer:
  version     : 0.1.0-rc6
  build date  : 2022-07-22T20:00:19Z
  git hash    : dac6a2a2edd891fe6dc338853878aac73c03ebc9
  go version  : go1.18.4
  go compiler : gc
  platform    : darwin/arm64
  features    : boilerplate=true

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

$ maintainer version
maintainer:
  version     : 0.1.0-rc7
  build date  : 2022-07-23T19:46:51Z
  git hash    : bf45f395772252f5592d0282e7aba51cd5239f22
  go version  : go1.18.4
  go compiler : gc
  platform    : darwin/arm64
  features    : boilerplate=true

$ maintainer github contribution suggest --delta 2022-05-01/+1         
 Day / Week                   #17      
----------------------- ---------------
 Sunday                        3       
 Monday                        4       
 Tuesday                       5       
 Wednesday                     6       
 Thursday                      6       
 Friday                        6       
 Saturday                      4       
----------------------- ---------------
 Suggestion is 2022-04-24: -92d, 3 → 6
```
