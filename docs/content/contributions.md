---
title: GitHub contributions
description: See your contribution pace and choose where to focus your next open source session.
---

# GitHub contributions

Use your contribution calendar as a record of visible activity: read recent weeks, identify a lighter day, and choose a target for your next open source session. These commands read the account identified by `GITHUB_TOKEN` or `--token` through GitHub's GraphQL API. They make no changes to your profile or repositories.

```sh
export GITHUB_TOKEN=your_token
maintainer github contribution lookup 2026-01-12/4
```

```text
 Day / Week   #01   #02   #03   #04   #05    Date
------------ ----- ----- ----- ----- ----- --------
 Sunday       25    25    25    15    15    Jan 25
 Monday       25    25    17    15    15    Jan 26
 Tuesday      25    25    15    15    15    Jan 27
 Wednesday    25    25    15    15    15    Jan 28
 Thursday     25    25    15    15    15    Jan 29
 Friday       25    25    15    15    15    Jan 30
 Saturday     25    25    15    15    15    Jan 31
------------ ----- ----- ----- ----- ----- --------
               distribution{15: 19, 17: 1, 25: 15}
```

The table lets you compare days across weeks. Weeks start on Sunday, and the week that contains January 1 is `#01`. A `-` is a day with no contributions; a `?` marks a future day in the displayed window. The final Date column locates each row in the last week. The footer counts the shown past days by their number of contributions: here 19 days had 15. Counts show GitHub activity, not the value or difficulty of the work. Without a token, the command fails before any request and names `GITHUB_TOKEN` and `--token`.

## See your recent pace

```sh
maintainer github contribution lookup                 # the latest commit's week and the one before
maintainer github contribution lookup now/-3          # recent weeks through today
maintainer github contribution lookup 2026-09-05/4    # centered on a date
maintainer github contribution lookup 2026-09-05/+2   # look forward from a date
```

The argument is `date[/weeks]`. The date is empty, `git`, `now`, a year (`2026`), a month (`2026-09`), a day (`2026-09-05`), or an RFC 3339 timestamp; a year or month is anchored at its first day, not shown whole. An empty date and `git` mean the author date of HEAD in the current repository, or in the one `GIT_DIR` points to; outside a repository, or in one without commits, they mean now. The weeks add to the date's week: `+N` after it, `-N` before it, or `N` split as ⌊N/2⌋ on each side. Without them, `lookup` uses `-1`. `N` is at most 53, so the table shows up to 54 weeks, the year is not before 1970, and a period entirely in the future is an error.

## Choose a contribution target

```sh
maintainer github contribution suggest --target 25 git/3
maintainer github contribution suggest --short --target 25 2026-01-12
git commit --date="$(maintainer github contribution suggest)"
```

`suggest` takes the same `date[/weeks]` argument, with `5` weeks by default; the date is the anchor. It answers with the first moment from the anchor up to now, within 05:00–19:00 UTC, on a day with fewer contributions than its week requires: the highest count in that week, but not less than `--target` (default `5`). The table goes to stderr with the day marked `*`, and stdout gets only the timestamp, so it can feed `git commit --date`. `--short` hides the table and `--delta` prints the time relative to now.

The moment includes a random jitter of less than an hour, so repeat runs differ, but it never moves past now or out of the schedule. If the anchor is in the future or no suitable moment exists up to now, `suggest` fails with a non-zero exit code instead of guessing. Nothing is committed or scheduled.

## Keep a baseline, then compare

```sh
maintainer github contribution snapshot 2026 > before.json
# Return later, after GitHub has recorded more contributions.
maintainer github contribution diff before.json 2026
```

`snapshot` writes a JSON object of UTC dates and daily counts for one calendar year, cut at now. Without a year it uses the current year. `diff` takes **base then head**: each argument is either a snapshot file or a four-digit year fetched from GitHub. It prints a row for each day that differs, with a signed difference:

```text
 Day          before   after   diff
------------ -------- ------- ------
 2026-09-08     25      28      +3
 2026-09-09     25      23      -2
```

A day only one side has counts as zero on the other and is shown as `-`, so it is listed only if it has contributions; a line under the table names the days only one side covers. When nothing differs, `diff` says `There is no diff between …`. For two saved points in time, run `maintainer github contribution diff before.json after.json`; comparing two files needs no token.

Snapshots record what GitHub reported at the time of the read. Current-day counts can still change as GitHub processes activity; keep the original file if you need a stable comparison.

## Find the shape of a year

```sh
maintainer github contribution histogram 2026
maintainer github contribution histogram 2026-09
maintainer github contribution histogram --with-zero 2026-09-25
```

`histogram` groups days by their contribution count and draws one `#` per day. Pass a year, month, or day; a day selects its week. Zero-count days are hidden by default, and `--with-zero` includes them.

## Useful boundaries

- A GitHub token is required when reading the live calendar, even for a public profile. The CLI queries the token owner's calendar rather than taking a username.
- `snapshot` takes one year, while `histogram` accepts year, month, or day. `lookup` and `suggest` accept date anchors and week spans.
- The `suggest` schedule is fixed at 05:00–19:00 UTC.
- Use `maintainer github contribution <command> --help` for the grammar and examples of each command.
