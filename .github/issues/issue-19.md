---
id: 19
database_id: 981576075
node_id: MDU6SXNzdWU5ODE1NzYwNzU=
status: open
title: "transition entity"
labels: [scope: code]
url: https://github.com/octomation/maintainer/issues/19
created_at: 2021-08-27T20:30:43Z
updated_at: 2023-03-31T15:34:35Z
---

# transition entity

Transition describes the transformation from state X to state Y. It is useful for dry-run and more accurate patch appling.

For example:

```
LabelX {
  name: "x"
  color: "000"
}

LabelY {
  name: "y"
  color: "000"
}

Transition {
  name: { from: "x", to: "y" }
}
```
