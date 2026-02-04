---
code: MAIN-159
id: I_kwDOE2M9Zc8AAAABdRbpLQ
databaseId: 6259403053
number: 521
url: https://github.com/octomation/maintainer/issues/521
title: "github: contribution: count the days as the profile does, wherever the account is"
issueType: Bug
labels:
  - "effort: medium"
  - "impact: high"
  - "scope: code"
  - "scope: docs"
  - "scope: test"
  - "severity: major"
  - "type: bug"
assignees:
  - kamilsk
milestone: "[[MAIN-M1]]"
parent: null
issueFields: {}
state: OPEN
stateReason: null
duplicateOf: null
createdAt: 2026-11-07T13:07:38Z
updatedAt: 2026-11-07T14:23:11Z
lastEditedAt: null
closedAt: null
---

# github: contribution: count the days as the profile does, wherever the account is

## Motivation

The profile shows one calendar, and `lookup` and `suggest` show another. For
kamilsk in 2025 both count 6141 contributions, but 160 days differ: January
3–5 are 15, 15, 15 on the profile and 0, 22, 23 in the commands. `suggest`
aims at the busiest day of a week, so one shifted spike raises the target of
the whole week.

The GraphQL `contributionCalendar` that [[MAIN-126]] chose puts a pull request
review on its date in the time zone of the token owner's account, and it moves
the whole history when that zone changes: Europe/Moscow on 2026-10-03, and
Asia/Tokyo after the owner moved. The profile puts reviews on their UTC date,
as GitHub documents for contributions in general. Commits, issues and pull
requests keep the same dates in both views. Neither the `Time-Zone` header, nor
the offset of `from` and `to`, nor the process time zone change it, and the API
does not tell the account's zone.

Separately, `suggest` prints the moment it chose in the machine's time zone, so
in Tokyo a late UTC moment becomes the author date of the next day.

## Expected behavior

- `lookup`, `suggest`, `histogram` and `snapshot` count every day of a calendar
  year as the profile's view of that year does, wherever the machine and the
  account are. The source stays the GraphQL API: the contribution events of each
  month of the year, placed by GitHub's rules:
  - a review on the UTC date it was submitted, one per pull request per year,
    the latest, as GitHub counts it;
  - commits on the days GitHub gives them;
  - opened issues, pull requests and created repositories on their UTC date.
- Fast: the events of each month are cached on disk. A run asks GitHub for the
  totals of all months of a year in one request and fetches only the months that
  are not cached or whose total has changed, so a run with nothing new is about
  as fast as one calendar request. A cache that cannot be read or written never
  fails a command: its months are fetched again.
- Reliable: an answer that does not add up is an error, not zeros. That covers a
  list that ends before its total, commits of a month in more than 100
  repositories, an unparseable date, and a GraphQL error.
- The API counts some contributions but gives no date for them, such as ones in
  private repositories the token cannot read. A month with them takes the days
  of the API calendar, which has every contribution: commits are on their days
  there, but a review may be a day off. If such a month has reviews, the
  command names it on stderr.
- `suggest` prints its moment in UTC (`Z`), so the author date keeps the chosen
  day in any time zone.

## Acceptance criteria

- [x] For kamilsk, every day of 2024, 2025 and 2026 up to the run agrees with the
  profile, with the machine in UTC, Europe/Moscow and Asia/Tokyo.
- [x] 2013, whose contributions are mostly in private repositories the token
  cannot read, agrees with the profile on every day; 2023 keeps the total of the
  API calendar and names the months whose reviews may be a day off.
- [x] A second `lookup` with no new contributions fetches no month and takes
  about as long as the calendar request did; a new contribution is seen on the
  next run; a removed or corrupt cache only makes that run slower.
- [x] `suggest` prints `Z`, and a late UTC moment stays on its day in Tokyo.
- [x] Unit tests cover the placement rules, the review per pull request, paging,
  incomplete answers, the cache and its invalidation; the integration tests pass
  against the live account.
- [x] The help, the user guide and the agent skill describe the source, the cache
  and the limits.

## Out of scope

- Parsing the profile HTML: [[MAIN-126]] keeps the API as the source.
- The profile's rolling view of the last twelve months: the commands count by
  calendar year.
- Placing a contribution the API gives no date for on the profile's day:
  the API calendar is the only source of its day.
- Issues opened in a browser keep the browser's date of the moment, and the
  API gives only the UTC moment. Such an issue opened near midnight may land one
  day off; none has so far.
- Account settings.

Related: [[MAIN-126]].
