---
id: 4
database_id: 777428833
node_id: MDU6SXNzdWU3Nzc0Mjg4MzM=
status: open
title: "how to warmup godoc and vanity url"
labels: []
url: https://github.com/octomation/maintainer/issues/4
created_at: 2021-01-02T08:55:25Z
updated_at: 2023-01-06T13:48:58Z
---

# how to warmup godoc and vanity url

when a new subpackage or package released

```bash
$ curl --data-urlencode 'path=go.octolab.org/toolkit/cli' \
  -H 'content-type: application/x-www-form-urlencoded' \
  https://godoc.org/-/refresh
```
