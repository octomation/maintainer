---
id: 90
database_id: 1522111801
node_id: I_kwDOE2M9Zc5auZU5
status: closed
title: "github: contribution: lookup call throw panic"
labels: [scope: code]
url: https://github.com/octomation/maintainer/issues/90
created_at: 2023-01-06T07:32:55Z
updated_at: 2023-01-06T13:48:01Z
---

# github: contribution: lookup call throw panic

Details

```
$ maintainer github contribution lookup /-2
panic: invalid count value:

goroutine 10 [running]:
go.octolab.org/toolset/maintainer/internal/service/github.contributionHeatMap.func1(0x14000355e18?, 0x1400037cc30)
	go.octolab.org/toolset/maintainer/internal/service/github/contribution.go:112 +0x2e0
github.com/PuerkitoBio/goquery.(*Selection).Each(0x1400037cc00, 0x14000355e48)
	github.com/PuerkitoBio/goquery@v1.8.0/iteration.go:10 +0x50
go.octolab.org/toolset/maintainer/internal/service/github.contributionHeatMap(0x140001240f0)
	go.octolab.org/toolset/maintainer/internal/service/github/contribution.go:108 +0x64
go.octolab.org/toolset/maintainer/internal/service/github.(*service).ContributionHeatMap.func1.1(0x100ba1a90?, {0x0?, 0x0?})
	go.octolab.org/toolset/maintainer/internal/service/github/contribution.go:44 +0xa8
go.octolab.org/toolset/maintainer/internal/service/github.(*service).ContributionHeatMap.func2()
	go.octolab.org/toolset/maintainer/internal/service/github/contribution.go:56 +0x48
golang.org/x/sync/errgroup.(*Group).Go.func1()
	golang.org/x/sync@v0.0.0-20220513210516-0976fa681c29/errgroup/errgroup.go:74 +0x60
created by golang.org/x/sync/errgroup.(*Group).Go
	golang.org/x/sync@v0.0.0-20220513210516-0976fa681c29/errgroup/errgroup.go:71 +0xa8
```
