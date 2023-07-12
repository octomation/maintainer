---
id: 59
database_id: 1273366858
node_id: I_kwDOE2M9Zc5L5glK
status: closed
title: "github: contribution: prevent caching of github contribution calendar"
labels: [scope: code]
url: https://github.com/octomation/maintainer/issues/59
created_at: 2022-06-16T10:25:03Z
updated_at: 2022-06-17T20:35:25Z
---

# github: contribution: prevent caching of github contribution calendar

**Motivation:** with `git contribute` it's possible to make a mistake based on stale data.

**Solution:** add `Cache-Control` header to prevent fetching cached responses.
