---
id: 60
database_id: 1274714555
node_id: I_kwDOE2M9Zc5L-pm7
status: closed
title: "github: contribution: replace --weeks by /weeks argument format for suggest"
labels: [scope: docs, scope: code]
url: https://github.com/octomation/maintainer/issues/60
created_at: 2022-06-17T08:08:26Z
updated_at: 2022-07-22T19:57:08Z
---

# github: contribution: replace --weeks by /weeks argument format for suggest

**Motivation:** make CLI consistent, see https://github.com/octomation/maintainer/blob/1a68d3ea715fd5b22500dc7a1c081fcca95784ad/internal/command/github/contribution.go#L188-L205.

**Interface**

```bash
$ maintainer github contribution suggest 2013-11-20/10
$ maintainer github contribution suggest 2013-11-20/+10
$ maintainer github contribution suggest 2013-11-20/-10
```
