---
id: 220
database_id: 2216519607
node_id: I_kwDOE2M9Zc6EHWe3
status: closed
title: "github: contributions: GitHub changed a way to provide data"
labels: ["scope: code","scope: test","type: bug","severity: critical","impact: high","effort: medium"]
url: https://github.com/octomation/maintainer/issues/220
created_at: 2024-03-30T15:09:41Z
updated_at: 2024-03-30T17:19:02Z
---

# github: contributions: GitHub changed a way to provide data

Restore contribution loading after GitHub moved to providing the calendar asynchronously. An ordinary profile page stopped being a sufficient source: the page HTML can load successfully and still contain none of the required cells.

The contract, illustrated:

```text
Request the chosen user and year
→ a response carrying the contributions calendar
→ the existing dates and counts in lookup/snapshot
```

The calendar content itself has to be fetched, and the same mechanism has to be used when the test data is refreshed. A successful HTTP response without the expected content must not be presented as a confirmed empty year.

**Current state:** the issue is closed. The [service](../../internal/service/github/contribution.go) uses a URL carrying `controller=profiles&action=show&tab=contributions` plus the year, and the [Taskfile](../../Taskfile) requests the same source — with an `X-Requested-With: XMLHttpRequest` header — when refreshing fixtures. This is a change distinct from moving the counts into a tooltip ([#174](issue-174.md)). Checking a real response on a schedule is related to [#219](issue-219.md).
