---
id: 81
database_id: 1337072715
node_id: I_kwDOE2M9Zc5PshxL
status: open
title: "makefile: build: allow include recursively"
labels: [scope: code, scope: test]
url: https://github.com/octomation/maintainer/issues/81
created_at: 2022-08-12T11:11:50Z
updated_at: 2023-08-09T12:38:09Z
---

# makefile: build: allow include recursively

**Motivation:** `build.service.mk` uses `build.tool.mk`. Now, this relation is implicit and resolved on the top level: `Go Service.mk`.
