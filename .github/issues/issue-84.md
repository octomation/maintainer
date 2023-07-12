---
id: 84
database_id: 1363105878
node_id: I_kwDOE2M9Zc5RP1hW
status: closed
title: "github: contribution: invalid suggestion for past"
labels: [scope: code, scope: test, type: bug]
url: https://github.com/octomation/maintainer/issues/84
created_at: 2022-09-06T11:01:52Z
updated_at: 2023-03-25T20:26:24Z
---

# github: contribution: invalid suggestion for past

**Steps to reproduce**

```
commit 9c834960956ad993a4a38e7f8a53a209511c53a6 (HEAD -> main)
Author: Kamil Samigullin <kamil@samigullin.info>
Date:   Thu Nov 11 13:57:51 2021 +0300

    build(deps-dev): bump @types/node from 18.7.14 to 18.7.15

commit d0e4a83b7cee98ae2ec89c2b6ab39b8b59793496 (origin/main, origin/HEAD)
Author: Kamil Samigullin <kamil@samigullin.info>
Date:   Fri Nov 12 13:51:50 2021 +0300

    build(deps-dev): bump @typescript-eslint/parser from 5.36.1 to 5.36.2

commit 17aba587e76fe9e782bf5989e743e7e9d8361c16
Author: Kamil Samigullin <kamil@samigullin.info>
Date:   Sun Nov 7 10:33:46 2021 +0300

    build(deps-dev): bump @typescript-eslint/parser from 5.35.1 to 5.36.1
```

```
maintainer github contribution suggest --delta 2021/+5
 Day / Week   #45   #46   #47   #48   #49   #50 
------------ ----- ----- ----- ----- ----- -----
 Sunday        9     6     2     -     8     5  
 Monday        9     -     -     -     4     1  
 Tuesday       9     -     6     1     -    10  
 Wednesday     9     1     2    11     2     8  
 Thursday      7     -     4     4     8     7  
 Friday        2     -     6     8     4     5  
 Saturday      4     -     -     -     8     1  
------------ ----- ----- ----- ----- ----- -----
 Suggestion is 2021-11-11: -299d, 7 → 9
```
