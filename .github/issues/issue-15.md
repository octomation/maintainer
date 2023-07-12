---
id: 15
database_id: 841075257
node_id: MDU6SXNzdWU4NDEwNzUyNTc=
status: closed
title: "extend octolab preset by other repositories"
labels: [scope: code]
url: https://github.com/octomation/maintainer/issues/15
created_at: 2021-03-25T15:59:21Z
updated_at: 2023-03-31T15:33:04Z
---

# extend octolab preset by other repositories

candidates:
- new subset
  - kamilsk/workshops
  - octolab/docs
  - octolab/rfc
- delete (aka empty subset)
  - kamilsk/gex
- update
  - kamilsk/bridge (from defaults)
  - kamilsk/breaker
  - kamilsk/check
  - kamilsk/dotfiles
  - kamilsk/egg
  - kamilsk/genome
  - kamilsk/grafaman

also add
- possibility to mix labels: `maintainer github labels patch octolab hacktoberfest`
- possibility to remove labels: `maintainer github labels patch empty`
