---
id: 78
database_id: 1330839082
node_id: I_kwDOE2M9Zc5PUv4q
status: open
title: "github: contribution: tips and tricks, daily snapshot"
labels: [scope: docs, type: feature, scope: inventory, impact: medium, effort: medium]
url: https://github.com/octomation/maintainer/issues/78
created_at: 2022-08-06T20:08:51Z
updated_at: 2023-04-06T11:14:35Z
---

# github: contribution: tips and tricks, daily snapshot

**Motivation:** there is no possibility to summarise your daily contribution impact.

**Interface**

```bash
$ cron 0 0 0 maintainer github contribution snapshot $(year - 1) $(year) > /tmp/daily.snapshot.json

$ maintainer github contribution diff progress
# maintainer github contribution diff --base=/tmp/daily.snapshot.json $(year - 1) $(year)
```
