---
code:
id: MDU6SXNzdWU4MjM2ODY5OTA=
databaseId: 823686990
number: 10
url: https://github.com/octomation/maintainer/issues/10
title: "template: automate sync with upstream"
labels: []
milestone:
state: OPEN
stateReason:
createdAt: 2021-03-06T16:24:19Z
updatedAt: 2023-10-31T08:45:34Z
lastEditedAt:
closedAt:
---

# template: automate sync with upstream

Automate updating a project created from a GitHub template once the template itself has changed. Today shared improvements to CI, the build and service files have to be carried over by hand, at the risk of overwriting adaptations specific to the project.

The source template and the comparison point need to be identified, the changes shown, and a proposal prepared for review. A prototype of the interface from the [idea draft](<../notes/Maintainer draft idea.md>):

```bash
maintainer layer template sync
# .github/workflows/ci.yml: update available
# Makefile: locally modified, needs review
```

This is a provisional interface: there is no `layer` command today. The criterion of success is a reviewable diff with a clear origin for every change; re-synchronizing an applied update must not offer it again. Conflicts are not resolved by silently replacing files.

One-off transfers of changes into maintainer itself are [#23](issue-23.md), [#69](issue-69.md), [#175](issue-175.md). Creating a new project from a template is the separate task [#153](issue-153.md).

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
