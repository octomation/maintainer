---
code: MAIN-184
id: I_kwDOE2M9Zc8AAAABigew2g
databaseId: 6610727130
number: 545
url: https://github.com/octomation/maintainer/issues/545
title: "service: hub: serve data only to the account that runs it"
issueType: Bug
labels:
  - "scope: docs"
  - "scope: code"
  - "scope: test"
  - "type: bug"
  - "severity: major"
  - "impact: high"
  - "effort: hard"
assignees:
  - kamilsk
milestone: null
parent: null
issueFields: {}
state: CLOSED
stateReason: COMPLETED
duplicateOf: null
createdAt: 2026-11-25T18:09:47Z
updatedAt: 2026-11-25T19:39:36Z
lastEditedAt: 2026-11-25T19:39:36Z
closedAt: 2026-11-25T18:44:04Z
---

# service: hub: serve data only to the account that runs it

## Problem

Another account of the same machine can read the data that the hub serves and
change the triage labels it writes. The hub must serve its data only to the
account that runs it.

## Expected behavior

Each run of the hub has its own origin whose name nobody can guess, and that
name is the credential; nothing is stored in the browser.

- At each start `service run` makes a label of 16 random bytes (32 hex
  characters) and serves its data only on `<label>.maintainer.localhost:<port>`.
  A restart makes a new label; the old origin is dead.
- For each port it serves, the hub writes the label to
  `$XDG_STATE_HOME/maintainer/service/<port>.label` (fallback
  `~/.local/state/…`): mode `0600`, folder `0700`, a temporary file and a
  rename. The label is in no log line and no error text. Only a clean stop
  removes the file, while the hub still holds its addresses, and only if the
  file still holds the label of this run. An error exit leaves the file; the
  next start replaces it. A start that fails writes and removes nothing.
- Before it writes the label or serves data, the hub holds both `127.0.0.1` and
  `[::1]` on each configured port. An address the machine does not have is
  skipped; an address another program holds stops the start with an error that
  names it.
- On the label host everything works as today, with the `Origin` and
  `X-Requested-With` checks for changes. On `maintainer.localhost`, `localhost`
  and the loopback literals only the pages, `/assets/` and
  `GET /api/v0/hello` answer; every other `/api/` route answers `401` with the
  code `unauthorized`. Any other host gets `421`. The label compare is
  constant-time.
- Nothing of the hub lives under a plain address: page requests on these
  hosts answer with `Clear-Site-Data: "storage"`, and on such an address the
  application also empties `localStorage` and `sessionStorage`.
- `GET /api/v0/hello?nonce=<hex>` returns `HMAC-SHA256(label, nonce)` as hex,
  so the command line can tell the user's hub from another listener.
- `maintainer service open` checks the proof of the hub with a fresh nonce,
  follows no redirect, and opens the label address in the browser without
  putting the label into any process argument. It says "the hub is not
  running" or "another program listens on the port" when the check fails.
  Before the label address, the browser passes through `maintainer.localhost`
  as a full page, so what earlier releases stored there is cleared; the label
  is in no request to the plain address. `--print` prints the URL to stdout
  instead.
- The application shows one full-page state for `401`, `421` or no answer:
  run `maintainer service open`. Before it shows it, the address drops its
  parameters. No owner, repository or issue name goes to `localStorage`,
  `sessionStorage` or a cookie; the theme may stay. No service worker.
- An open tab asks `hello` every 30 seconds while it is visible, and when it
  becomes visible again, so it shows the state within about 30 seconds after
  the hub stops.
- The protection is on by default and can be turned off on a machine where
  the user is the only account: `access = "open"` in `service.toml` makes the
  plain addresses serve everything, as before this change, and keep their
  browser storage. The `Host`, `Origin` and `X-Requested-With` checks stay.
  The label, its file and `maintainer service open` keep working. The start
  says in its log that access is open. Any value other than `"label"` (the
  default) or `"open"` is a configuration error.

## Acceptance criteria

- [x] On `maintainer.localhost`, `localhost` and `127.0.0.1`, each `/api/v0/`
      route except `hello` answers `401` and returns no data; on the label host
      of this run it works as before; the label of an earlier run gets `421`.
- [x] The label file has mode `0600`; a restart changes the label; a clean stop
      removes the file and an error exit leaves it; a second instance that
      fails to bind leaves the file of the first as it is.
- [x] With `listen = ["127.0.0.1:<port>"]` the hub also holds `[::1]:<port>`;
      if another program holds that address, the start fails, names it, and no
      label file appears.
- [x] `maintainer service open --print` prints a URL that opens the hub with
      data; against a listener that does not know the label the command
      refuses; no process argument of the command or of the browser launcher
      holds the label.
- [x] Page requests on `maintainer.localhost`, `localhost` and `127.0.0.1`
      answer with `Clear-Site-Data: "storage"`; after a page load there the
      browser storage of that address is empty.
- [x] The page that `maintainer service open` gives the browser passes
      through `maintainer.localhost` before the label address; afterwards the
      storage of `maintainer.localhost` is empty, and no request to it holds
      the label.
- [x] The application stores no owner, repository or issue name in
      `localStorage`, `sessionStorage` or a cookie, and registers no service
      worker.
- [x] When the hub stops, an open tab shows the state within about 30 seconds,
      without a user action, and its address has no parameters.
- [x] The caveats that Homebrew prints name `maintainer service open`.
- [x] `docs/content/service.md` has a section on access with the limits below;
      the agent skill reference names the command.
- [x] With `access = "open"`, `maintainer.localhost`, `localhost` and
      `127.0.0.1` serve every route with data, answer without
      `Clear-Site-Data`, and the application keeps the theme there; another
      host still gets `421`; the start log says that access is open; any other
      value of `access` stops the start with a configuration error.
- [x] `docs/content/service.md` describes the setting and when to use it.
- [x] `make check` and the `web/` gates pass; `web/dist` is rebuilt.

## Out of scope

- HTTPS with a local certificate.
- A process that runs as the user can read the label file.
- Limits stated in the guide: after each start of the hub, run
  `maintainer service open`; the label is in the user's own browser history;
  the connection is plain HTTP, so a tab left open after the hub stops keeps
  requesting its address; `maintainer service open` clears only
  `maintainer.localhost`, and only with JavaScript on: the user opens any
  other plain address once; the theme resets at each start.

Source: [the security review of 2026-11-25](../reports/20261125T152238Z.md), F4.
