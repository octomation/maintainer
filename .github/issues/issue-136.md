---
id: 136
database_id: 1676018941
node_id: I_kwDOE2M9Zc5j5gT9
status: open
title: "github: contribution: incorrect suggest for today"
labels: [scope: code, type: bug, severity: critical, impact: high, effort: easy]
url: https://github.com/octomation/maintainer/issues/136
created_at: 2023-04-20T05:19:35Z
updated_at: 2023-04-20T09:18:36Z
---

# github: contribution: incorrect suggest for today

**Details**

```bash
$ git contrib docs: readme: improve headline

 Day / Week   #15   #16    Date  
------------ ----- ----- --------
 Sunday       10    10    Apr 16 
 Monday       10    10    Apr 17 
 Tuesday      10    10    Apr 18 
 Wednesday    10     8    Apr 19 
 Thursday     10     *    Apr 20 
 Friday       10     ?    Apr 21 
 Saturday     10     ?    Apr 22 
------------ ----- ----- --------
              Stats: coming soon 

Suggestion is , 0 → 10
info  - Loaded env from /Users/ksamigullin/Development/public/tact-app/web/.env
✔ No ESLint warnings or errors
[main 902e583] docs: readme: improve headline
 Date: Thu Apr 20 08:29:24 2023 +0300
 2 files changed, 6 insertions(+), 6 deletions(-)



$ git --no-pager log -2
commit 902e5839c6640c007ab6b5c265f9580b63607dae (HEAD -> main)
Author: Kamil Samigullin <kamil@samigullin.info>
Date:   Thu Apr 20 08:29:24 2023 +0300

    docs: readme: improve headline

commit c3854cbea22b489c263b13505f78cf5e64be9f7a (origin/main, origin/HEAD)
Author: dependabot[bot] <49699333+dependabot[bot]@users.noreply.github.com>
Date:   Wed Apr 19 23:25:37 2023 +0300

    tools(deps): bump vercel from 28.18.5 to 28.19.0 in /tools (#722)
```
