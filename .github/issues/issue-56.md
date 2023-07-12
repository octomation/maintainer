---
id: 56
database_id: 1271291420
node_id: I_kwDOE2M9Zc5Lxl4c
status: closed
title: "github: contribution: support --short for suggest"
labels: [scope: docs, scope: code]
url: https://github.com/octomation/maintainer/issues/56
created_at: 2022-06-14T19:54:39Z
updated_at: 2022-06-15T10:51:18Z
---

# github: contribution: support --short for suggest

**Motivation:** improve usage experience with Git.

**PoC**

```bash
$ git config alias.contribute="!git at $(maintainer github contribution suggest --short $$1) $${1:}"
$ git contribute 2021 some message
```
