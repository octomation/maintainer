---
code: MAIN-126
id: null
database_id: null
node_id: null
status: open
title: "github: contribution: calendar counts are read from an inconsistent source"
labels: ["scope: code","scope: test","scope: deps","type: bug","severity: major","impact: high","effort: medium"]
url: null
created_at: 2026-09-01T11:11:10Z
updated_at: 2026-09-01T11:11:10Z
---

# github: contribution: calendar counts are read from an inconsistent source

**Motivation:** every contribution command reports counts that disagree with the
profile page, and re-running the same command changes the answer. The value a
user acts on depends on which cache the request happened to reach, so the only
way to see the current number is to invoke the command repeatedly until it
stops moving.

**Details**

The heat map is scraped from the year-scoped variant of the profile page:

```
https://github.com/<user>?controller=profiles&action=show&tab=contributions&from=<year>-01-01
```

That variant is served from several caches populated at different moments.
Six consecutive requests to this exact URL returned three different values for
the day still being indexed, while all other 364 days were identical in every
response:

```
snap 1: 2026-08-31=24  total=5092
snap 2: 2026-08-31=22  total=5090
snap 3: 2026-08-31=24  total=5092
snap 4: 2026-08-31=19  total=5087
snap 5: 2026-08-31=19  total=5087
snap 6: 2026-08-31=22  total=5090

differing days: 2026-08-31 ['24','22','24','19','19','22']
```

The inconsistency is server-side and cannot be worked around by the client:

- the responses carry `cache-control: max-age=0, private, must-revalidate` with
  no `age` and no `x-cache`, so the existing `Cache-Control`, `Pragma` and
  `Expires` request headers have nothing to invalidate;
- adding a unique query parameter per request does not stabilise the answer;
- `X-Requested-With: XMLHttpRequest` is load-bearing for a different reason —
  without it the page returns an `include-fragment` placeholder and the calendar
  is absent from the response entirely.

Three other sources answer consistently, and agree with each other:

| source | 8 requests | verdict |
| --- | --- | --- |
| `?tab=contributions&from=<year>-01-01` | 19 / 22 / 24 | inconsistent |
| `?tab=contributions` (what the browser loads) | 24 ×8 | consistent |
| `/users/<login>/contributions?from=&to=` | 24 ×8 | consistent |
| GraphQL `contributionsCollection` | 24 ×8 | consistent |

GraphQL is the only one of them that is a documented API rather than a page the
project happens to parse, and its day buckets are independent of the offset
passed in `from`/`to` (`Z`, `+03:00` and `-07:00` return identical dates).
It caps a single collection at one year, which matches the per-year fan-out the
service already performs.

Required behaviour:

- a day's count never walks backwards between two reads of the same source;
- the calendar is read through the GraphQL API, not by parsing profile HTML;
- the per-year fan-out is preserved, and each query spans exactly one calendar
  year so the one-year cap is never exceeded;
- a GraphQL failure reported as `200 OK` with a non-empty `errors` array is
  surfaced as an error, not silently decoded as an empty calendar;
- an unauthorized or throttled client fails with the transport status rather
  than an empty heat map — GraphQL requires a token even for public profiles;
- the HTML parser and the dependencies that exist only to serve it are removed
  once nothing reads HTML.

**PoC**

Against a day that is still being indexed:

```bash
$ for i in $(seq 1 6); do
    curl -s -H "Authorization: Bearer $GITHUB_TOKEN" \
         -H 'X-Requested-With: XMLHttpRequest' \
         "https://github.com/$USER?tab=contributions&from=$(date +%Y)-01-01" \
      | grep -o '>[0-9]* contributions on August 31st\.'
  done
>19 contributions on August 31st.
>19 contributions on August 31st.
>22 contributions on August 31st.
>24 contributions on August 31st.
>19 contributions on August 31st.
>22 contributions on August 31st.
```

The same window through the API does not move:

```bash
$ maintainer github contribution snapshot | jq '."2026-08-31T00:00:00Z"'
```

The regression test should read the current year twice and assert that no day
decreased, so a source that serves aged snapshots fails it while genuine
indexing of new contributions does not.

<!--
## What was done

The investigation started from a mismatch: the command reported 19 contributions
for August 31st while the profile page showed 23. The first guess, an HTTP cache,
did not hold up — the response headers forbid caching and carry neither `age` nor
`x-cache`. What settled it was diffing six consecutive responses to the same URL:
exactly one day disagreed, the other 364 matched. So it is neither our cache nor
a timezone, but several server-side snapshots of different ages.

Checked separately that cache busting is useless: a unique query parameter does
not change the picture. `X-Requested-With` must stay for an unrelated reason —
without it the calendar is absent from the response entirely and the page returns
an `include-fragment` placeholder.

`internal/service/github/graphql.go`: a minimal client over the existing oauth2
`http.Client`, no new dependencies. The case where GraphQL answers `200 OK` with
a non-empty `errors` array is handled explicitly; otherwise a failure would decode
into an empty calendar.

`FetchContributions` now returns `contribution.HeatMap` instead of
`*goquery.Document`. `ContributionHeatMap` and the `Contributor` interface are
unchanged, so the commands and the mocks were not touched. The year bounds are
`YYYY-01-01T00:00:00Z … YYYY-12-31T23:59:59Z`: verified that a future `to` is
accepted and the range comes back exactly as requested, without padding to whole
weeks.

`BuildHeatMap`, `YearRange` and the three packages that existed only to serve the
scraper — `internal/pkg/url`, `internal/pkg/http/header`, `internal/pkg/errors` —
became dead and were removed; `goquery` left `go.mod`. The model HTML fixtures
were regenerated into JSON by the same parser, in the format the `snapshot`
command already writes, so the data and the expected values in the tests stayed
the same.

Tests: `graphql_test.go` is hermetic, an `httptest` server behind a redirecting
transport (decoding, the `from`/`to` bounds, three classes of error). The
integration `TestService_FetchContributions_Consistency` reads the current year
six times and fails on any decrease; growth is allowed, that is genuine indexing.

Verified live: six consecutive `contribution snapshot` runs produced an identical
slice. The old endpoint had converged by then — its caches caught up — so the
jitter only reproduces while a day is still being indexed. Worth keeping in mind
before trying to reproduce the observation after the fact.

Left over: `contribution lookup` still panics on an assertion and so took no part
in the verification — split out into [[MAIN-125]].

Reopened the next day: `suggest` reported 21 for September 1st while the profile
showed 22, and the binary was the GraphQL build. Four identical whole-year
queries answered 21, 22, 20, 20 — the very same jitter, now behind
`contributionsCollection`. Every other range was steady at 22 across five
calls: a short one, a fixed `to` a day ahead, and `to` equal to the current
second. So it is not the endpoint that is cached but the query: a range GitHub
has seen before comes back from whichever cache answers, while a range nobody
asked for before is resolved from the source. The whole-year range is the most
popular key there is.

The fix bounds the current year by the present second, which makes every call
a new key; past years keep the fixed bounds, their data is immutable. The
service got a `now` clock so the hermetic test can pin the `to` value. Six
consecutive `contribution lookup` runs agreed on 22 afterwards.
-->
