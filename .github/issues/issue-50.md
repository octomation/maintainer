---
id: 50
database_id: 1263735611
node_id: I_kwDOE2M9Zc5LUxM7
status: open
title: "command: separate command structure and command execution"
labels: [scope: code, scope: test, type: improvement, impact: medium, effort: medium]
url: https://github.com/octomation/maintainer/issues/50
created_at: 2022-06-07T18:59:04Z
updated_at: 2023-04-06T06:13:34Z
---

# command: separate command structure and command execution

**Motivation:** it allows mock dependencies. Now it's difficult because the configuration is available only in runtime after a command call.

**PoC:**

```go
func Runner(deps ...) func(cobra.Command, args) error {}
```
