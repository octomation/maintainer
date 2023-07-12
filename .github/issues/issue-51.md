---
id: 51
database_id: 1263738579
node_id: I_kwDOE2M9Zc5LUx7T
status: open
title: "di: define service provider to inject it into command for lazy service initialization"
labels: [scope: code, scope: test]
url: https://github.com/octomation/maintainer/issues/51
created_at: 2022-06-07T19:01:10Z
updated_at: 2023-03-31T15:34:36Z
---

# di: define service provider to inject it into command for lazy service initialization

**Motivation:** see #50.

**PoC**:

```go
func RunE(provider interface { SomeDeps() Service }) func(cobra.Command, args) error {}
```

**To do:**

- [ ] check https://github.com/google/wire
- [ ] check https://github.com/uber-go/fx
