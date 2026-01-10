---
code:
id: I_kwDOE2M9Zc5tCLYi
databaseId: 1829287458
number: 153
url: https://github.com/octomation/maintainer/issues/153
title: "template: go new support"
labels:
  - "scope: code"
  - "type: feature"
  - "impact: high"
  - "effort: hard"
milestone:
state: OPEN
stateReason:
createdAt: 2023-07-31T14:14:42Z
updatedAt: 2023-08-26T19:26:36Z
lastEditedAt: 2023-08-26T19:22:14Z
closedAt:
---

# template: go new support

Simplify creating a project from a template: get a ready structure with a new module name and consistent imports, without copying files by hand or running a mass search and replace. The original idea combines Go templates with GitHub template repositories, and aims to port the approach to other stacks for a better developer experience.

The reference from the [Go blog post](https://go.dev/blog/gonew) is a separate `gonew` tool rather than a built-in `go new` command:

```bash
gonew golang.org/x/example/helloserver example.com/myserver
```

The proposed result for maintainer: choose the template source and version, name the new project, and get consistent files plus clear next steps. An existing non-empty directory must not be overwritten implicitly. Support for other stacks should be advertised according to the adapters that actually exist, not promised universally in advance.

**At present** no such command exists. Renaming Go paths is [#135](issue-135.md); synchronizing the created project with its template later is [#10](issue-10.md). Additional original links: [tools v0.11.1 release](https://github.com/golang/tools/releases/tag/v0.11.1), [GitHub templates](https://docs.github.com/en/repositories/creating-and-managing-repositories/creating-a-template-repository).

<!-- 2023-08-09T12:36Z https://github.com/octomation/maintainer/issues/153#issuecomment-1671245480
related to #135
-->

<!-- 2023-08-26T19:26Z https://github.com/octomation/maintainer/issues/153#issuecomment-1694486616
- [ ] investigate how does it change a module name https://github.com/golang/tools/blob/master/cmd/gonew/main.go
-->
