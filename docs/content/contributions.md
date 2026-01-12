---
title: GitHub contributions
description: See your contribution pace and choose where to focus your next open source session.
---

# GitHub contributions

Use your contribution calendar as a record of visible activity: read recent weeks, identify a lighter day, and choose a target for your next open source session. These commands read the account identified by `GITHUB_TOKEN` or `--token` through GitHub's GraphQL API. They make no changes to your profile or repositories.

```sh
export GITHUB_TOKEN=your_token
maintainer github contribution lookup now/-3
```

The table lets you compare days across weeks. A `-` is a day with no contributions; a `?` marks a future day in the displayed window. The final Date column locates each row in the range. Counts show GitHub activity, not the value or difficulty of the work.

## See your recent pace

```sh
maintainer github contribution lookup now/-3          # recent weeks through today
maintainer github contribution lookup 2026-09-25/4    # centered on a date
maintainer github contribution lookup 2026-09-25/-4   # look backward
maintainer github contribution lookup git/4           # centered on this repo's latest commit
```

`lookup` accepts a year (`2026`), month (`2026-09`), day (`2026-09-25`), RFC 3339 timestamp, `now`, or `git` as an anchor. Add `/N` for a centered span, `/+N` for weeks forward, or `/-N` for weeks backward. `git` uses the latest commit author's date in a Git checkout and falls back to now elsewhere.

**For this version, always include a nonzero week span with `lookup`.** The bare command and date-only forms can panic; they are tracked in [issue #155](https://github.com/octomation/maintainer/issues/155). A year or month acts as an anchor, not as a whole-year or whole-month display.

## Choose a contribution target

```sh
maintainer github contribution suggest --target 25 now/-3
maintainer github contribution suggest --short --target 25 now/-3
maintainer github contribution suggest --delta git/5
```

`suggest` highlights a candidate day with `*` and reports its current count and effective target. `--target` sets a minimum (default `5`); the effective target rises to the highest count in the candidate week when that is larger. `--short` omits the table for scripts. `--delta` prints a relative time too. The chosen time includes randomness, so repeat runs can differ. This is planning advice: nothing is committed or scheduled.

The `git` anchor uses the latest commit author's date, so it can choose a past date. Use `now` when planning from the present.

## Keep a baseline, then compare

```sh
maintainer github contribution snapshot 2026 > before.json
# Return later, after GitHub has recorded more contributions.
maintainer github contribution diff before.json 2026
```

`snapshot` writes a JSON object of UTC dates and daily counts for one calendar year. Without a year it uses the current year. `diff` takes **base then head**: each argument is either a snapshot file or a four-digit year fetched from GitHub. The table displays positive changes. For two saved points in time, run `maintainer github contribution diff before.json after.json`.

Snapshots record what GitHub reported at the time of the read. Current-day counts can still change as GitHub processes activity; keep the original file if you need a stable comparison. **A decrease currently wraps to a huge positive number** in `diff` because counts use unsigned arithmetic. Check the two JSON files directly if a decrease is possible ([issue #70](https://github.com/octomation/maintainer/issues/70)).

## Find the shape of a year

```sh
maintainer github contribution histogram 2026
maintainer github contribution histogram 2026-09
maintainer github contribution histogram --with-zero 2026-09-25
```

`histogram` groups days by their contribution count and draws one `#` per day. Pass a year, month, or day; a day selects its week. Zero-count days are hidden by default, and `--with-zero` includes them.

## Useful boundaries

- A GitHub token is required when reading the live calendar. The CLI queries the token owner's calendar rather than taking a username.
- `snapshot` takes a year, while `histogram` accepts year, month, or day. `lookup` and `suggest` accept date anchors and week spans.
- A `diff` queries GitHub if either argument is a year; comparing two files needs no token or calendar fetch.
- Use `maintainer github contribution <command> --help` to inspect the installed version's flags.
