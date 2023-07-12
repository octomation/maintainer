---
id: 193
database_id: 2071900858
node_id: I_kwDOE2M9Zc57frK6
status: open
title: "github: suggestion: reduce jitter"
labels: [scope: code, type: improvement, impact: high, effort: easy]
url: https://github.com/octomation/maintainer/issues/193
created_at: 2024-01-09T08:49:32Z
updated_at: 2024-01-09T08:49:33Z
---

# github: suggestion: reduce jitter

**Motivation:** it limits available capacity for the day.

**Details**

https://github.com/octomation/maintainer/blob/020d048eebf8059814ca342b4ccd1e34a87784c1/internal/command/github/contribution/suggest.go#L50

**To do**

I must change it to a "target-specific" value but limit it to top.
