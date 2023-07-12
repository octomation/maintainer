---
id: 38
database_id: 1235562560
node_id: I_kwDOE2M9Zc5JpTBA
status: closed
title: "github: contribution: lookup shows incorrect scope for 1 week with now ts"
labels: [scope: code]
url: https://github.com/octomation/maintainer/issues/38
created_at: 2022-05-13T18:34:22Z
updated_at: 2022-06-15T10:15:49Z
---

# github: contribution: lookup shows incorrect scope for 1 week with now ts

```bash
$ maintainer github contribution lookup /1
 Day / Week                                    #18
---------------------------------- ---------------------------
 Sunday                                         6
 Monday                                         6
 Tuesday                                        6
 Wednesday                                      1
 Thursday                                       6
 Friday                                         6
 Saturday                                       6
---------------------------------- ---------------------------
 Contributions are on the range from 2022-05-01 to 2022-05-07
```

`#19` needs to be shown

```bash
maintainer github contribution lookup /2
 Day / Week               #17           #18           #19
-------------------- ------------- ------------- -------------
 Sunday                    3             6             6
 Monday                    4             6             4
 Tuesday                   5             6             6
 Wednesday                 6             1             6
 Thursday                  6             6             6
 Friday                    6             6             4
 Saturday                  4             6             ?
-------------------- ------------- ------------- -------------
 Contributions are on the range from 2022-04-24 to 2022-05-13
```

`#17` needs to be unshown
