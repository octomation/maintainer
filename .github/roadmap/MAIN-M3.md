---
code: MAIN-M3
id: MI_kwDOE2M9Zc4BFBVP
databaseId: 18093391
number: 3
url: https://github.com/octomation/maintainer/milestone/3
title: "Hub, v0.3"
state: OPEN
openIssueCount: 5
closedIssueCount: 1
dueOn: null
createdAt: 2026-09-26T09:14:44Z
updatedAt: 2026-10-21T10:24:10Z
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
- [x] [[MAIN-155]] — service: hub: runtime for the pages at maintainer.localhost
- [x] [[MAIN-156]] — service: status: the checkout table in the browser
- [x] [[MAIN-157]] — service: inbox: the issues of the collection as a mail app
- [x] [[MAIN-158]] — service: tablo: the issues of the collection as a kanban board
