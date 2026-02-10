---
code: MAIN-10
id: MDU6SXNzdWU4MjM2ODY5OTA=
databaseId: 823686990
number: 10
url: https://github.com/octomation/maintainer/issues/10
title: "template: automate sync with upstream"
issueType: null
labels: []
assignees:
  - kamilsk
milestone: "[[MAIN-M1]]"
parent: null
issueFields: {}
state: CLOSED
stateReason: NOT_PLANNED
duplicateOf: null
createdAt: 2021-03-06T16:24:19Z
updatedAt: 2026-11-21T05:33:42Z
lastEditedAt: 2026-11-20T15:42:13Z
closedAt: 2026-11-21T05:33:42Z
---

# template: automate sync with upstream

## Motivation

A project created from a GitHub template stops following it on day one. Shared
improvements to CI, the build and the service files then have to be carried
over by hand, and a careless copy overwrites what the project adapted on
purpose. maintainer itself went through this more than once: [[MAIN-23]],
[[MAIN-69]] and [[MAIN-116]] are one-off transfers from go-tool.

The idea is to automate it: remember the template commit a project was created
from, compute the upstream diff since that checkpoint, and suggest the changes
for review. The checkpoint was meant to live in a GitHub
[custom property](https://docs.github.com/en/rest/repos/properties?apiVersion=2022-11-28#about-custom-properties)
of the repository. The first templates to support are go-module, go-tool and
go-service; the reference case outside them is withsparkle/obsidian, created
from obsidianmd/obsidian-sample-plugin at its commit 7112f01.

## PoC

```bash
$ maintainer layer template sync
# template:  octomation/go-tool, checkpoint a1b2c3d → e4f5a6b
# .github/workflows/ci.yml: update available
# Makefile: locally modified, needs review
```

## Acceptance criteria

- [ ] The template and the checkpoint are found for a project, from the
  repository's record or given explicitly; an unknown one is a clear error.
- [ ] The result is a reviewable diff in which every change names its template
  commit.
- [ ] A file the project changed is reported for review, never silently
  replaced.
- [ ] After an update is applied, the checkpoint moves, and the next run does
  not offer the same change again.
- [ ] go-module, go-tool and go-service are covered by tests.

## Out of scope

- Creating a project from a template: [[MAIN-112]].
- Resolving merge conflicts automatically.

Related: [[MAIN-23]], [[MAIN-69]], [[MAIN-116]], [[MAIN-112]].

<!-- 2023-10-31T08:43Z https://github.com/octomation/maintainer/issues/10#issuecomment-1786756606
it should be based on https://docs.github.com/en/rest/repos/properties?apiVersion=2022-11-28#about-custom-properties

Algorithm:
- create from template
- set checkpoint value
- calculate upstream diff based on checkpoint
- suggest changes
-->

<!-- 2023-10-31T08:45Z https://github.com/octomation/maintainer/issues/10#issuecomment-1786758465
Example

![image](https://github.com/octomation/maintainer/assets/1165416/1376b2c1-c75b-4b56-857b-47db114b637a)

- https://github.com/withsparkle/obsidian uses https://github.com/obsidianmd/obsidian-sample-plugin
- the repository was created on 7112f01 commit
-->

<!-- 2023-10-31T08:45Z https://github.com/octomation/maintainer/issues/10#issuecomment-1786759542
add support to
- [ ] go-module
- [ ] go-tool
- [ ] go-service
-->

<!-- 2026-11-20T21:00Z https://github.com/octomation/maintainer/issues/10#issuecomment-7043710704
Maturity: idea. Nothing in code: there is no template or layer command (internal/command/root.go), and nothing records a checkpoint. Fetch discovery reads only whether a repository is itself a template (`IsTemplate` in internal/service/github/discovery_rest.go), not which template it came from.
Proposal for v0.5.0: not planned. The job is merging template files into a project's tree, which is boilerplate maintenance like the removed build helpers, and copier or cruft already do it.
A read-only signal fits the mission and could be a separate task: which checkouts of the workspace lag their template, shown by status or the hub.
-->

<!-- 2026-11-21T05:30Z https://github.com/octomation/maintainer/issues/10#issuecomment-7057974961
Closed as not planned: this is no longer the focus of maintainer. From v0.5.0 maintainer is the contributor's central tool: one picture of the repositories, issue trackers and GitHub across all projects, not build or release tooling.
The need itself stays: keeping a project in sync with its template is recorded on #153, together with copier and cruft.
-->
