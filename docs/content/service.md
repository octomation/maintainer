---
title: Local service
description: One local hub for your GitHub work at http://maintainer.localhost:17841/ — every checkout and every issue your repositories mirror from GitHub, in the status, inbox and tablo pages.
---

# Local service

`maintainer service run` gives your development on GitHub one place on your
computer: a hub at `http://maintainer.localhost:17841/` with every checkout of
the collection that [`maintainer fetch`](/fetch/) and
[`maintainer status`](/status/) know about, and every issue those
repositories mirror from GitHub in `.github/issues` and `.github/roadmap`, with
their milestones and project-board sections, across all your owners. Instead
of opening repositories one by one, you see in three pages where local work
waits, which issues matter most and where each one stands, with a link to
GitHub when the mirror has one. The tool that keeps the mirror holds those
files in step with GitHub; the hub itself never reads from or writes to
GitHub. It needs no token or network access, runs as your user and listens
only on loopback: `.localhost` names resolve to your own computer, so there is nothing
to set up and no administrator step.

| Page | What it shows |
| --- | --- |
| `/status` | The checkouts of `maintainer status` in the browser, grouped by their state: branch, uncommitted changes, divergence, push lock and path per checkout, with the same ordering and search |
| `/inbox` | Every mirrored issue in one list with the open issue beside it, filters, a score and the `impact` and `effort` labels you set in place as its priority and estimate |
| `/tablo` | The same issues on one board with the columns Backlog, To do, In progress, Done and Won't do |

![The hub at maintainer.localhost: the inbox with an open issue in front of the status page](/images/hub/hero.webp)

*Illustrations of the hub on demo data.*

## Install and start

On macOS, the Homebrew Formula `octolab/tap/maintainer` installs the CLI and
the service definition; the service is the same binary as the CLI and
upgrades with it.

```sh
brew install --formula octolab/tap/maintainer
brew services start octolab/tap/maintainer
curl -fsS http://maintainer.localhost:17841/status -o /dev/null -w '%{http_code}\n'
```

```text
200
```

Then open the hub:

```sh
maintainer service open
```

It opens the status page in your browser at the private address of the
running service ([access](#access)). Run it again after every start of the
service. The service runs as your user and logs to
`$(brew --prefix)/var/log/maintainer.log`. Installing the Formula does not
start it. Do not run `sudo brew services`.

## Move from the Cask

If you installed the Cask, switch to the Formula once:

```sh
brew uninstall --cask maintainer && brew install --formula octolab/tap/maintainer
```

Your fetch configuration and state stay where they are.

## Upgrade, stop and remove

```sh
brew update
brew upgrade --formula octolab/tap/maintainer
brew services restart octolab/tap/maintainer
brew services stop octolab/tap/maintainer
brew uninstall --formula octolab/tap/maintainer
```

## Run it yourself

Without Homebrew, run the service in a terminal:

```sh
maintainer service run
```

```text
maintainer service: 127.0.0.1:17841, [::1]:17841 (URL: http://maintainer.localhost:17841/)
maintainer service: run `maintainer service open` to open it in the browser
maintainer service: fetch config /Users/you/.config/maintainer/fetch.toml
```

Then run `maintainer service open` in another terminal.

`maintainer.localhost` resolves to both `127.0.0.1` and `::1`, so the service
listens on both. If the system has no IPv6 loopback, the service says so,
skips `[::1]:17841` and keeps running on `127.0.0.1:17841`. `--listen`
replaces both addresses; repeat it for several. For each port, the service
also listens on the other of `127.0.0.1` and `::1`, so these two lines start
the same service:

```sh
maintainer service run --listen 127.0.0.1:8080
maintainer service run --listen 127.0.0.1:8080 --listen '[::1]:8080'
```

Each value is a literal loopback IP address with a port, such as
`127.0.0.1:17841` or `[::1]:17841`; `localhost`, `0.0.0.0` or another address
is refused with exit code `2`. An address that cannot be bound, as when
another program holds its port on either `127.0.0.1` or `::1`, stops the
service with an error that names it. `Ctrl-C` or `SIGTERM` stops it once the
responses in flight are sent.

`brew services` runs the service without flags. To move it to another port
there, set `listen` in [`service.toml`](#the-config-file) and restart it with
`brew services restart octolab/tap/maintainer`; the same rules apply, and
`--listen` replaces `listen` when you run the service yourself:

```toml filename="service.toml"
listen = ["127.0.0.1:8080", "[::1]:8080"]
```

Browsers and `curl` resolve `.localhost` names themselves; a browser that
does not cannot open the hub. On another port,
pass the address to `maintainer service open --listen 127.0.0.1:8080`, or set
`listen` in `service.toml`, which the command reads too.

## Access

The hub shows your issues and the paths of your checkouts, private ones
included, and changes labels in your issue files, so only you get its data.
Every start of the service makes a new random label and serves the data only
at `http://<label>.maintainer.localhost:17841/`. The label is the key: the
service writes it to `$XDG_STATE_HOME/maintainer/service/17841.label`
(`~/.local/state/maintainer/service/` by default), readable only by you, and
removes the file when it stops cleanly. The browser keeps nothing of the hub but the
theme.

`maintainer service open` reads the label, checks that the service on the
port knows it, and opens the address in your browser:

```sh
maintainer service open
maintainer service open --print
```

```text
http://3f9c0e8a1b2d4c6e8f0a1b2c3d4e5f60.maintainer.localhost:17841/
```

`--print` prints the address instead of opening it. `--listen` names the
address of a service on another port; without it, the command takes the first
address of `listen` in `service.toml`, else `127.0.0.1:17841`. When the
service is not running, the command says so; when another program answers on
the port, it says that and opens nothing.

The plain address `http://maintainer.localhost:17841/` still loads the pages,
and `curl …/status` still prints `200`, but they show no data: the page says
to run `maintainer service open`. The same page appears in a tab of an
earlier run after the service restarts or stops. The plain address keeps
nothing in the browser: its pages clear the browser storage of that address,
and `maintainer service open` passes through it on the way to the address of
the run, so the view state that earlier releases kept there is removed.

Limits:

- A process that runs as your user can read the label file. The hub does not
  protect you from your own programs.
- After each start of the service, such as an upgrade, a reboot or
  `brew services restart`, run `maintainer service open`; a bookmark of the
  plain address shows the page that says so.
- A program that takes the port while the service is down can show a false
  page on the plain address. It gets no data and no label.
- The label is in the address bar and the history of your browser.
- The connection is plain HTTP, so the browser cannot tell the hub from a
  program that takes the port after the hub stops. Such a program sees the
  addresses that an open tab still requests: the label of the ended run, the
  page, filters and issue ids. On a computer with other accounts, close the
  tabs of the hub before you stop it.
- The theme resets at each start of the service, because the address is new.
- `maintainer service open` clears what earlier releases stored under
  `http://maintainer.localhost:17841/` only, and only with JavaScript on. If
  JavaScript is off, or if you ever opened the hub through
  `http://localhost:17841/` or an IP address, open each such address once
  after the upgrade: the page clears what earlier releases stored there.

### Open access

The label protects the hub from the other accounts of your computer. If you
are its only account, you can turn it off in [`service.toml`](#the-config-file)
and use the plain address as before v0.5.1:

```toml filename="service.toml"
access = "open"
```

Then restart the service. `http://maintainer.localhost:17841/`,
`http://localhost:17841/` and the loopback addresses of the port serve the
data, a bookmark and an open tab keep working after a restart, and the
browser keeps the theme. The service says so when it starts:

```text
maintainer service: access is open: the plain address serves the data too (access = "open" in service.toml)
```

Other web sites still get nothing: the service answers only its own host
names and accepts changes only from its own pages. `maintainer service open`
works as before. Any other account of the computer can read the data, so
keep the default, `access = "label"`, on a shared computer.

## What the pages read

![How it fits: GitHub mirrors issues into .github/issues of your checkouts, the hub at maintainer.localhost reads them, brew services runs it and service.toml configures it](/images/hub/how-it-works.webp)

The service reads the fetch configuration that `[fetch] config` in
`service.toml` names; otherwise it looks, as `maintainer status` does, in
`$MAINTAINER_FETCH_CONFIG`, then in `$XDG_CONFIG_HOME/maintainer`
(`~/.config/maintainer` by default), and never in the working directory. A
`fetch.toml` or `fetch.yaml` in your home folder is not read: the service logs
a notice, and without another fetch configuration it does not start. With
the state, it finds the checkouts of the [workspace](/workspace/), the pins
included.
Every checkout that has a `.github/issues` or `.github/roadmap` directory is an
issue mirror: the service reads the issue files, the milestones and the
boards into an index when it starts, and again when you press **Reindex** in
the sidebar; a label you set from a page is in the index at once. A file it
cannot read is reported on the page and keeps its last good version. The pages
open in the dark theme; the theme item at the bottom of the sidebar switches
to light or to your system's setting.

The sidebar also lists every repository whose issues the hub reads, as a tree
by owner: an owner expands to its repositories, each with its count of open
issues and a lock when it is private, and a link to its issues in the inbox.
The owners you expand stay open while the page is open, and the repository
the inbox shows is highlighted. Its search
button, or `/` or Ctrl+K (Cmd+K on a Mac), opens a search over every issue of
the collection; the issue you pick opens as a full page, as does an issue you
open from the board or expand from the inbox. The page's address carries the
issue, so a link or a reload opens it again; **Back** or Escape returns to
the list. **Filter** (or the F key) narrows the inbox and the board by state,
text, priority, estimate, labels, type, milestone, repository, owner or
assignee; the inbox also filters by status. **View** groups the inbox by status, priority, estimate,
repository, owner, milestone or type; a group's count covers every issue of
the selection, and **Load more** fills the groups. On the board, **View**
orders the cards of every column
by score, priority, or the updated or created date (by default the open
columns by score and Done and Won't do by the last update), hides the empty
columns and chooses what a card shows.

![The inbox grouped by priority next to an issue opened as a full page with its properties and score](/images/hub/inbox.webp)

The status page collects the Git state of the checkouts on its first visit
and on **Refresh** in its **View** menu, which also shows the summary of
the collection and the time of the snapshot. Ordering and searching never run
Git again: **View** groups the checkouts by state or by owner, sets the
ordering, Shift-click adding a column to it, and chooses the columns of the
list (branch, changes, status, path), which line up from row to row;
**Filter** searches the checkouts by repository, branch and path, and the
list/board switch shows the groups as board columns.

![The status page: checkouts grouped by state with aligned columns, and the repository tree with locks on private repositories](/images/hub/status.webp)

An issue goes to a board column by its state: a closed issue is in **Done**,
or in **Won't do** when it was closed as not planned or as a duplicate; an
open issue is in the column of the board section that lists it, when the
section is called To do, In progress or Backlog, and in **Backlog** otherwise.
An open issue that two boards place differently stays in Backlog with a
marker; an open issue that a board lists under Done or Won't do gets a
contradiction marker, and another board's open section can still decide its
column.

![The board with issues of several repositories in five columns and the search open over it](/images/hub/tablo.webp)

## Triage and the score

The inbox and the board let you set two labels on an issue, `impact: high`,
`medium` or `low` and `effort: easy`, `medium` or `hard`: the issue shows
`impact` as its **Priority** (High, Medium, Low) and `effort` as its
**Estimate** (S, M, L), and picking another value in either menu sets the
label, or removes it with No priority or No estimate. The service changes
only those label lines in the issue file, keeps everything else byte for
byte, and refuses the change when the file changed since the page read it:
reload the issue and set the label again. Pushing the labels to GitHub stays
with the tooling that maintains the mirror. Do not run that tooling's export
or push, or another editor of the issue files, while you triage in the
browser.

![The priority menu open on an issue page: No priority, High, Medium, Low](/images/hub/triage.webp)

In the inbox, **View**, then **Ordering**, sorts the issues by a score, by
priority (the `impact` label), or by the updated or created date. The score is a weighted sum that a repository
weight multiplies:

```text
score = repository weight × (impact × w_impact + ease × w_effort + type × w_type + age × w_age + due × w_due)
```

Each input is between 0 and 1: `impact: high` is 1, `medium` 0.5, `low`
0.25; `effort: easy` is 1, `medium` 0.5, `hard` 0.25; a bug is 1, a feature
0.5; the age grows with the days since the issue was created and reaches 1
after the age horizon; the due term starts rising the due horizon before the
milestone's due date and is 1 from the due date on. A missing label or date
counts as 0. The inbox shows the terms of every score next to the issue.

The weights, the horizons and the repository weights come from
[`service.toml`](#the-config-file); by default every weight is `1`, the age
horizon is 365 days, the due horizon 30 days, and every repository weighs `1`.
For example:

```toml filename="service.toml"
# [fetch]
# config = "~/.config/maintainer/fetch.toml"  # the fetch config to read; the usual lookup otherwise

[score]
impact = 2
effort = 1
type = 0.5
age = 1
due = 3
age_horizon_days = 180
due_horizon_days = 14

[[repos]]
match = { owner = "octomation", name = "maintainer" }
weight = 2

[[repos]]
match = { id = 672356038 }  # a repository id survives renames and transfers
weight = 0.5
```

A repository rule matches by `id`, or by `owner` and `name` together, without
regard to case; the first matching rule wins.

## The config file

The service reads `service.toml` in `$XDG_CONFIG_HOME/maintainer`, else
`~/.config/maintainer`. When there is none, the first start writes it there,
says so, and never changes it again: every key is commented out with its
default, so the file shows where the service listens and what you can change.

```text
maintainer service: wrote the default config /Users/you/.config/maintainer/service.toml
```

Uncomment a key to change it, then restart the service. `--config` names
another file, which must exist and is never created. The file is strict: an
unknown key or a bad value is an error, and the service does not start. A
relative path in it is resolved against your home directory.

## What it answers

| Request | Response |
| --- | --- |
| `GET` or `HEAD` of `/` | `302` to `/status` |
| `GET` or `HEAD` of `/status`, `/inbox`, `/tablo` | `200`, the application as `text/html`, never cached (`Cache-Control: no-store`) |
| `GET` of `/assets/…` | the application's script and styles, cached for a year |
| `GET` of `/api/v0/…` | the JSON the pages read: `checkouts`, `tasks`, `tasks/{id}`, `catalog`; only at the address of the run |
| `GET` of `/api/v0/hello?nonce=…` | the proof of the label that `maintainer service open` checks |
| `PATCH` of `/api/v0/tasks/{id}/triage`, `POST` of `…/refresh` | a change or a refresh, only from the pages themselves |
| any other path | `404 Not Found` |
| any other method | `405 Method Not Allowed` with `Allow` |

The service shows the contents of your issue files, so it answers only
requests addressed to itself. The address of the run gets everything.
`maintainer.localhost`, `localhost` and the address it listens on get the
pages, the assets and `hello`; every other `/api/` request there gets `401`
with the code `unauthorized`. Any other `Host`, the address of an earlier run
included, gets `421`. A change or a refresh is accepted only from a page the
service served, never from another site open in your browser.

## Limitations

- The pages work on the mirror files; nothing is read from or written to
  GitHub.
- The index is rebuilt on start and on **Reindex**; a file you edit by hand
  shows up after the next reindex.
- The board has one fixed set of columns; the boards of your projects are
  read for the column of an issue, not shown as they are.
- Linux has no service manager integration yet: run `maintainer service run`
  yourself.
