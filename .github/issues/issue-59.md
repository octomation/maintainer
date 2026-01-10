---
code:
id: I_kwDOE2M9Zc5L5glK
databaseId: 1273366858
number: 59
url: https://github.com/octomation/maintainer/issues/59
title: "github: contribution: prevent caching of github contribution calendar"
labels:
  - "scope: code"
milestone: "[[milestone-1]]"
state: CLOSED
stateReason: COMPLETED
createdAt: 2022-06-16T10:25:03Z
updatedAt: 2022-06-17T20:35:25Z
lastEditedAt: 2022-06-17T20:26:09Z
closedAt: 2022-06-17T20:35:25Z
---

# github: contribution: prevent caching of github contribution calendar

Reduce the risk of choosing a date from a stale GitHub calendar. Across successive calls of a Git wrapper, a cached response can suggest a day that has already been filled.

An example of the scenario: get a suggestion → publish the contribution → repeat the request. A fresh request is expected, one that does not rely on a stored response from the client or an intermediate cache. The solution was to send `Cache-Control` headers that prevent cached responses from being used.

**Current state:** the issue is closed; [loading contributions](../../internal/service/github/contribution.go) sends no-cache headers. That reduces the influence of the HTTP cache, but it does not guarantee that GitHub itself recounts contributions immediately. When describing the outcome, the freshness of the request must be distinguished from the freshness of the source data.

Snapshots stored by the user keep their meaning unchanged — that is the separate scenario [#27](issue-27.md).

<!-- 2022-06-17T20:35Z https://github.com/octomation/maintainer/issues/59#issuecomment-1159210651
https://github.com/octomation/maintainer/commit/5c984d531a67f78eee1ad97e2f356dfa758c8f2b
-->
