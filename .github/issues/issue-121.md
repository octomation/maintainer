---
id: 121
database_id: 1641609817
node_id: I_kwDOE2M9Zc5h2PpZ
status: closed
title: "github: contribution: lookup command is corrupted"
labels: [scope: code, scope: test, type: bug, severity: major, impact: medium, effort: easy]
url: https://github.com/octomation/maintainer/issues/121
created_at: 2023-03-27T08:09:12Z
updated_at: 2023-03-29T13:41:39Z
---

# github: contribution: lookup command is corrupted

**Details**

```bash
$ maintainer github contribution lookup 2022/+10
Error: please provide argument in format YYYY-mm-dd[/[+|-]weeks], e.g., 2006-01-02/3: invalid argument "2022/+10": parsing time "2022" as "2006-01-02": cannot parse "" as "-"
```

Affected release: https://github.com/octomation/maintainer/releases/tag/v0.1.0-rc10.
