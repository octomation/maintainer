---
id: 89
database_id: 1501357396
node_id: I_kwDOE2M9Zc5ZfOVU
status: closed
title: "git: config: contribution since"
labels: [scope: code]
url: https://github.com/octomation/maintainer/issues/89
created_at: 2022-12-17T12:04:38Z
updated_at: 2023-03-25T20:55:08Z
---

# git: config: contribution since

**Motivation:** to separate "start time" per project. It allows me to avoid the manual job.

For example,
- I've been developing the Tact app since May 1, 2022
- But other projects since 2021

**PoC**

```bash
# inside some Tact's repo
$ git config contribution.since 2022-05-01
$ git contribute
# maintainer suggest contribution $(git config contribution.since || git latest commit)
```
