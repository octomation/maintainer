---
id: 43
database_id: 1253612392
node_id: I_kwDOE2M9Zc5KuJto
status: closed
title: "github: contribution: expand heat map by merging with neighbors"
labels: []
url: https://github.com/octomation/maintainer/issues/43
created_at: 2022-05-31T09:45:24Z
updated_at: 2022-06-15T10:22:32Z
---

# github: contribution: expand heat map by merging with neighbors

Now heat map truncated by year, because GitHub sliced them on this way. But, it will be great to expand it, e.g.

```bash
$ maintainer github contribution lookup 2013-12-31/5
 Day / Week                #51            #52           #1
--------------------- -------------- -------------- ----------
 Sunday                     -              -            -
 Monday                     -              -            -
 Tuesday                    -              2            -
 Wednesday                  -              2            ?
 Thursday                   4              -            ?
 Friday                     3              2            ?
 Saturday                   -              -            ?
--------------------- -------------- -------------- ----------
 Contributions are on the range from 2013-12-15 to 2013-12-31
```

```bash
maintainer github contribution lookup 2014-01-01/5
 Day / Week                #1            #2            #3
--------------------- ------------- ------------- ------------
 Sunday                     ?             -            -
 Monday                     ?             -            3
 Tuesday                    ?             -            -
 Wednesday                  -             -            2
 Thursday                   -             -            -
 Friday                     -             -            -
 Saturday                   -             -            -
--------------------- ------------- ------------- ------------
 Contributions are on the range from 2014-01-01 to 2014-01-18
```

So, `/5` doesn't work, it is truncated by year, from left for 2014, and from right for 2013. Also, the result contains `?` which means no data here.


- [ ] remove `TrimByYear`
```go
			scope := xtime.
				RangeByWeeks(date, weeks, half).
				Shift(-xtime.Day).
				ExcludeFuture().
				TrimByYear(date.Year())
```
- [ ] extend to support many years
```go
chm, err := service.ContributionHeatMap(cmd.Context(), date)
```
