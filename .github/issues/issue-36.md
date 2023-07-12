---
id: 36
database_id: 1234170490
node_id: I_kwDOE2M9Zc5Jj_J6
status: closed
title: "command: github contribution lookup doesn't work well with now()"
labels: [scope: code, scope: test]
url: https://github.com/octomation/maintainer/issues/36
created_at: 2022-05-12T15:38:44Z
updated_at: 2022-06-15T10:15:49Z
---

# command: github contribution lookup doesn't work well with now()

expected

![image](https://user-images.githubusercontent.com/1165416/168114309-beb71f1a-c88f-4221-8cd8-1e56c2ab4453.png)

obtained

```
$ maintainer github contribution lookup /2
 Day / Week               #17           #18           #19
-------------------- ------------- ------------- -------------
 Sunday                    -             -             -
 Monday                    -             -             -
 Tuesday                   -             -             -
 Wednesday                 -             -             -
 Thursday                  -             -             -
 Friday                    -             -             ?
 Saturday                  -             -             ?
-------------------- ------------- ------------- -------------
 Contributions are on the range from 2022-04-24 to 2022-05-12
```
