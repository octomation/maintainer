---
code: MAIN-160
id: I_kwDOE2M9Zc8AAAABdSIojQ
databaseId: 6260140173
number: 522
url: https://github.com/octomation/maintainer/issues/522
title: "github: contribution: lose no contribution at the edges of the monthly slices"
issueType: Task
labels:
  - "scope: code"
  - "scope: test"
  - "type: improvement"
  - "impact: medium"
  - "effort: medium"
assignees:
  - kamilsk
milestone: "[[MAIN-M1]]"
parent: null
issueFields: {}
state: OPEN
stateReason: null
duplicateOf: null
createdAt: 2026-11-07T14:23:34Z
updatedAt: 2026-11-07T18:51:49Z
lastEditedAt: 2026-11-07T18:51:49Z
closedAt: null
---

# github: contribution: lose no contribution at the edges of the monthly slices

## Motivation

[[MAIN-159]] builds the days of a year from the contribution events of each
month, fetched as separate `contributionsCollection` slices. GitHub selects the
events of a slice by whole days in the time zone of the request, so the slices
of a year fit together only while every request of a run is placed in the same
zone. If GitHub mixed contexts between the requests of one run, the events at
the edge of two months could be counted twice or not at all, and nothing would
tell. No run has shown it so far (15 of 15 identical requests on 2026-10-31
agreed), but identical calendar requests did alternate between two zones in an
earlier session.

The last slice of the current year has an edge of its own: it ends now, but
GitHub ends it with today in the zone of the request. With the account in
Tokyo, today ends at 15:00Z, and a review after that is in tomorrow, which no
slice asks for. On 2026-11-07 the profile had 17 contributions on that day,
and `lookup` and `suggest` had 5 until the next UTC day. The total of the month
did not change either, so the cache looked fresh.

## Scope

- The current year reaches now in any zone: an event of today after the end
  of today in the zone of the request counts on its day.
- The slices of a year overlap at their edges, and an event that two slices
  return is counted once.
- A run that still cannot tell whether an edge event belongs to the year is an
  error or a named warning, never a silent change of a day.

## Acceptance criteria

- [x] A review of today after the end of today in the zone of the request,
  such as one at 17:56Z with the account in Tokyo, counts today on the next
  run, and the commands agree with the profile on that day.
- [ ] With the requests of one run placed in different zones, every day of the
  year keeps the count of the profile.
- [ ] An event returned by two slices is counted once, including the commits of
  one repository on one day and the review of one pull request.
- [ ] Warm runs stay about as fast as before: the overlap costs no request when
  the months are cached.
- [ ] Unit tests cover the mixed zones and the duplicates; the integration tests
  pass against the live account.

## Out of scope

- Contributions the API gives no date for: their months keep the days of the
  API calendar, as [[MAIN-159]] decides.
- Parsing the profile HTML.

Related: [[MAIN-159]], [[MAIN-126]].
