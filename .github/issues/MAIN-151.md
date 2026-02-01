---
code: MAIN-151
id: null
databaseId: null
number: null
url: null
title: "service: serve maintainer locally at maintainer.localhost"
issueType: Feature
labels:
  - "type: feature"
  - "scope: code"
  - "scope: test"
  - "scope: docs"
  - "impact: medium"
  - "effort: medium"
assignees:
  - kamilsk
milestone: "[[MAIN-M2]]"
parent: null
issueFields: {}
state: OPEN
stateReason: null
duplicateOf: null
createdAt: null
updatedAt: null
lastEditedAt: null
closedAt: null
---

# service: serve maintainer locally at maintainer.localhost

## Motivation

maintainer works in the terminal, but some of what it knows reads better in a browser on the same computer; the next such page is a board of the tasks across the user's repositories. maintainer needs a local service for it, run by `brew services` as the user, at `http://maintainer.localhost:17841/`: `.localhost` names resolve to loopback without any setup, so no root, hosts entry or port redirect is involved. indexit has the same kind of service on the next port down, and both follow one shape. The first version serves the smallest content, a plain page that says the service is running, so that installation, upgrades and removal are settled before any data is shown.

## PoC

```bash
$ maintainer service run
maintainer service: 127.0.0.1:17841, [::1]:17841 (URL: http://maintainer.localhost:17841/)
$ curl -fsS http://maintainer.localhost:17841/ | grep -o 'Service is running.'
Service is running.

$ brew install --formula octolab/tap/maintainer
$ brew services start octolab/tap/maintainer
$ curl -fsS http://maintainer.localhost:17841/ -o /dev/null -w '%{http_code}\n'
200
```

## Expected behavior

- `maintainer service run` serves HTTP on loopback only. Without flags it listens on `127.0.0.1:17841` and `[::1]:17841`, because `maintainer.localhost` resolves to both; 17840 belongs to indexit, and the ports of local services are recorded in AGENTS.md. `--listen` is repeatable and replaces the defaults; each value must be a literal loopback IP address, IPv4 or IPv6, with a port, and anything else is a usage error, exit `2`.
- When one default address cannot be bound because the system lacks it, as when IPv6 is turned off, the service runs on the other and says so; when none can, or an address given with `--listen` fails, it stops with an error.
- The first version answers `GET` and `HEAD` of `/` with a plain HTML page that says "Service is running.": `Content-Type: text/html; charset=utf-8`, `X-Content-Type-Options: nosniff`, `Cache-Control: no-store`. Other paths are `404`, other methods `405` with `Allow: GET, HEAD`.
- `SIGINT` and `SIGTERM`, which `brew services stop` sends, stop it gracefully: responses in flight finish, then it exits with `0`.
- It runs as the user, never as root, and needs no configuration, token or network access.
- The Homebrew Formula ([[MAIN-125]]) declares the service: `brew services start octolab/tap/maintainer` runs `maintainer service run` with its defaults from the same binary as the CLI and restarts it if it stops.
- A guide describes the installation, the check, upgrades, stopping and removal, `--listen`, and the move from the Cask.

## Acceptance criteria

- [ ] `maintainer service run --listen 127.0.0.1:0 --listen '[::1]:0'` serves the page at `/` on both addresses with the headers above; `--listen 0.0.0.0:17841` and `--listen localhost:17841` fail with exit `2`.
- [ ] Without flags the service listens on `127.0.0.1:17841` and `[::1]:17841`; a default address the system cannot assign is skipped with a message, an explicit one is an error.
- [ ] `httptest` covers the routes, the methods, `HEAD`, the conditional request and a graceful shutdown with a request in flight.
- [ ] The generated Formula has a `service do` block that runs `maintainer service run`, and its caveats name `http://maintainer.localhost:17841/` and `brew services start octolab/tap/maintainer`.
- [ ] After `brew services start octolab/tap/maintainer`, `curl -fsS http://maintainer.localhost:17841/ -o /dev/null -w '%{http_code}\n'` prints `200`.

## Out of scope

The task board and any other page or data; a service manager on Linux; HTTPS; access control beyond loopback. Before the service shows private data, it must also reject requests whose `Host` is not its own, against DNS rebinding.
