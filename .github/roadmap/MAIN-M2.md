---
code: MAIN-M2
id: MI_kwDOE2M9Zc4Ae1gM
databaseId: 8083468
number: 2
url: https://github.com/octomation/maintainer/milestone/2
title: "Hub, v0.3"
state: OPEN
openIssueCount: 4
closedIssueCount: 1
dueOn: null
createdAt: 2022-06-15T10:13:13Z
updatedAt: 2026-10-20T07:43:09Z
closedAt: null
---

# Hub, v0.3

## Maintainer hub

A web interface to manage repositories and tasks in one place: local, fast, and synchronized with the trackers where the tasks live. `maintainer` serves it on the loopback through `brew services`, so it opens instantly and keeps working offline.

The v0.3 line supports one backend, GitHub. Linear, Jira and other trackers are the direction the hub grows in, behind the same model of tasks; they are not part of v0.3.

### Value

1. One place for the collection and its work — the checkouts that `fetch` and `status` know about, next to the issues, milestones and boards that concern them.
2. Local and fast — the hub works on local state and synchronizes with the backend, instead of a browser round trip to each tracker.
3. One model of tasks, ready for more trackers than GitHub.

### Scope of v0.3.0

- The local service at `http://maintainer.localhost:17841/`, started with `brew services`.
- One Homebrew Formula for the CLI and the service; the Cask is retired.
- Repositories and GitHub tasks in the hub, synchronized with GitHub.

## Issues

- [ ] [[MAIN-25]] — add possibility to compare configuration of repos
- [ ] [[MAIN-31]] — command: github backup repository
- [x] [[MAIN-53]] — tools: experimental: investigate cli dashboard for GitHub
- [ ] [[MAIN-105]] — linear: generate client
- [ ] [[MAIN-125]] — release: homebrew: ship one formula with the service and retire the cask
- [ ] [[MAIN-151]] — service: serve maintainer locally at maintainer.localhost
