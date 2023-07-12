---
id: 65
database_id: 1294048701
node_id: I_kwDOE2M9Zc5NIZ29
status: closed
title: "github: contribution: lookup has problem with timezone"
labels: [scope: code, scope: test]
url: https://github.com/octomation/maintainer/issues/65
created_at: 2022-07-05T09:38:50Z
updated_at: 2022-07-05T10:34:15Z
---

# github: contribution: lookup has problem with timezone

```bash
$ maintainer github contribution lookup /-5
 Day / Week      #22     #23     #24     #25     #26     #27
-------------- ------- ------- ------- ------- ------- -------
 Sunday           5       5       5       5       5       3
 Monday           4       5       3       5       5       6
 Tuesday          5       4       5       5       4       3
 Wednesday        5       5       5       5       4       ?
 Thursday         4       4       3       4       4       ?
 Friday           5       5       4       5       4       ?
 Saturday         5       4       3       5       2       ?
-------------- ------- ------- ------- ------- ------- -------
 Contributions are on the range from 2022-05-29 to 2022-07-05
```

But on the same time

<img width="260" alt="image" src="https://user-images.githubusercontent.com/1165416/177298172-c6fb76df-e451-450f-ac27-97124eeca477.png">
