---
id: 98
database_id: 1535840228
node_id: I_kwDOE2M9Zc5biw_k
status: open
title: "github: setup: add command to manage all available projects and repositories"
labels: [scope: code, type: feature, impact: high, effort: hard]
url: https://github.com/octomation/maintainer/issues/98
created_at: 2023-01-17T06:23:16Z
updated_at: 2023-01-17T06:23:16Z
---

# github: setup: add command to manage all available projects and repositories

**Motivation:** configure forks and upstream automagically.

**Algorithm**

- if it's fork: configure `upstream`
- if it has forks: configure remotes: `fork-%owner`
- fetch changes through the whole dir structure
- init dir structure tree and manage it, shows stats above it
