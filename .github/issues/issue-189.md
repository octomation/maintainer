---
id: 189
database_id: 2068634275
node_id: I_kwDOE2M9Zc57TNqj
status: open
title: "github: contribution: suggest doesn't work properly in headless mode"
labels: [scope: code, type: bug, severity: major, effort: medium]
url: https://github.com/octomation/maintainer/issues/189
created_at: 2024-01-06T13:33:42Z
updated_at: 2024-01-06T13:33:43Z
---

# github: contribution: suggest doesn't work properly in headless mode

**Details**

If I use something like this `export GIT_DIR=path/to/.git`, it cannot define the correct suggestion.

E.g.,

- correct suggestion

```bash
$ maintainer github contribution suggest git/1

 Day / Week   #52    Date
------------ ----- --------
 Sunday       50    Dec 24
 Monday       50    Dec 25
 Tuesday      50    Dec 26
 Wednesday    50    Dec 27
 Thursday     50    Dec 28
 Friday       45*   Dec 29
 Saturday      -    Dec 30
------------ ----- --------
        Stats: coming soon

Suggestion is 2023-12-29T16:33:16+03:00, 45 → 50
```

- incorrect suggestion

```bash
maintainer github contribution suggest git/1

 Day / Week   #53    Date
------------ ----- --------
 Sunday        -    Dec 31
 Monday       15    Jan  1
 Tuesday      15    Jan  2
 Wednesday     8    Jan  3
 Thursday      5    Jan  4
 Friday        9    Jan  5
 Saturday     13*   Jan  6
------------ ----- --------
        Stats: coming soon

Suggestion is 2024-01-06T16:42:05+03:00, 13 → 15
```

P.S.: I have to check it with `git --git-dir=path/to/.git ...`.
