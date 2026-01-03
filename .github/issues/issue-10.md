---
id: 10
database_id: 823686990
node_id: MDU6SXNzdWU4MjM2ODY5OTA=
status: open
title: "template: automate sync with upstream"
labels: []
url: https://github.com/octomation/maintainer/issues/10
created_at: 2021-03-06T16:24:19Z
updated_at: 2023-10-31T08:45:34Z
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
