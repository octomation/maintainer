---
id: 133
database_id: 1666111192
node_id: I_kwDOE2M9Zc5jTtbY
status: open
title: "github: contribution: suggest for cmm"
labels: [scope: code, scope: test, type: bug, severity: major, impact: medium, effort: easy]
url: https://github.com/octomation/maintainer/issues/133
created_at: 2023-04-13T10:02:01Z
updated_at: 2023-04-13T10:02:02Z
---

# github: contribution: suggest for cmm

**Details**

```bash
$ git contrib ...
# add commit for future
$ git contrib ...
recovered: assertion is not a true
---
unexpected panic occurred
# because last commit was for the future
```

```bash
$ git log
commit 1e6a5a23aa36a7c239c62352df868f3afa1d90d8 (HEAD -> main)
Author: Kamil Samigullin <kamil@samigullin.info>
Date:   Thu Apr 13 13:06:20 2023 +0300

    fix #694: docs: integrate Nextra for docs publishing

commit d0c41cc9a42622ef9312ec7394d7d5ee30b96ca9 (tag: v0.3.0-pre.2, origin/main, origin/HEAD)
Author: Kamil Samigullin <kamil@samigullin.info>
Date:   Thu Apr 13 12:34:13 2023 +0300

    docs: changelog: describe v0.3.0-pre.2
```

So, it must be limited by `now()`.
