---
id: 128
database_id: 1656731036
node_id: I_kwDOE2M9Zc5iv7Wc
status: open
title: "github: contribution: support preset for suggestion"
labels: [scope: code, scope: test, type: feature, effort: medium]
url: https://github.com/octomation/maintainer/issues/128
created_at: 2023-04-06T06:11:45Z
updated_at: 2023-04-06T06:11:45Z
---

# github: contribution: support preset for suggestion

**Motivation:** reduce counter distribution.

```go
func Suggest(
	heats HeatMap,
	scope xtime.Range,
	hours xtime.Schedule,
-	basis unit,
+	basis []unit,
) Suggestion {
```

Defaults:  `[]uint{5, 10, 15}`.
