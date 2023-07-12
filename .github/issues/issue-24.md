---
id: 24
database_id: 1216404100
node_id: I_kwDOE2M9Zc5IgNqE
status: closed
title: "use contributions chart for git at suggestion"
labels: [scope: code]
url: https://github.com/octomation/maintainer/issues/24
created_at: 2022-04-26T19:47:24Z
updated_at: 2022-06-15T10:15:47Z
---

# use contributions chart for git at suggestion

research
- https://github.com/sallar/github-contributions-chart
- https://github.com/Bloggify/github-calendar

acceptance criteria:
- algorithm finds holes
- defines "minimum" commits
- suggests date for `git at -123d`
- allows to view contributions chart without github

```bash
$ git at $(maintainer contribution suggest --for=2021) did something useful
```
