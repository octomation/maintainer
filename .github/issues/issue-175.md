---
id: 175
database_id: 2000333790
node_id: I_kwDOE2M9Zc53Oqve
status: open
title: "maintainance: sync with the go-tool template"
labels: ["effort: easy"]
url: https://github.com/octomation/maintainer/issues/175
created_at: 2023-11-18T09:14:11Z
updated_at: 2023-11-18T09:14:11Z
---

# maintainance: sync with the go-tool template

Synchronize maintainer's infrastructure with go-tool so that the checks and the publishing match the environment the project actually supports. The motivation came from a [workflow failure](https://github.com/octomation/maintainer/actions/runs/6912316252/job/18807924831); the original report carries no error text, so its exact cause is not asserted here.

The result must include a reviewable diff of the template settings and must preserve maintainer's own scenarios, above all the daily check of the GitHub calendar. Mechanically copying every template file is not enough.

**Current context:** `go.mod` requires Go 1.26, while the workflows still name 1.21.x and 1.22.x explicitly; the toolset and the release build commands have also changed separately. The advertised support, the toolchain actually selected, the generation and the delivery all have to be brought into agreement — rather than declaring some old version number the proven cause of a historical failure.

Readiness means the mandatory jobs are verified and the intentional differences from the chosen template revision are understood. General automation of synchronization is [#10](issue-10.md), the earlier release preparation is [#69](issue-69.md), and the workflow-management context is [this note](<../notes/Issues/work with workflows.md>).
