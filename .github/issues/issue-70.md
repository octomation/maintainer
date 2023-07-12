---
id: 70
database_id: 1315648314
node_id: I_kwDOE2M9Zc5OazM6
status: open
title: "github: contribution: refactor diff command"
labels: [scope: docs, scope: code, scope: test, type: improvement, impact: medium, effort: medium]
url: https://github.com/octomation/maintainer/issues/70
created_at: 2022-07-23T12:31:24Z
updated_at: 2023-04-06T11:15:57Z
---

# github: contribution: refactor diff command

**Motivation:** it contains invalid logic and showed problem after refactoring after [1522818](https://github.com/octomation/maintainer/commit/1522818c8dba3194378ffb83088be880b0245c32).

```bash
$ maintainer github contribution diff --base=/tmp/snap.01.2013.json --head=/tmp/snap.02.2013.json
 Day / Week                  #46             #48             #49           #50    
---------------------- --------------- --------------- --------------- -----------
 Sunday                       -               -               -             -     
 Monday                       -               -               -             -     
 Tuesday                      -               -               -             -     
 Wednesday                   -4               -              -1             -     
 Thursday                     -               -               -            -1     
 Friday                       -              -2               -             -     
 Saturday                     -               -               -             -     
---------------------- --------------- --------------- --------------- -----------
 The diff between head{"/tmp/snap.02.2013.json"} → base{"/tmp/snap.01.2013.json"}
```

must be `+` instead of `-`.

Also, `--base=/tmp/snap.01.2013.json --head=/tmp/snap.02.2013.json` should be args, not params. See `diff` as reference, https://linuxize.com/post/diff-command-in-linux/.
