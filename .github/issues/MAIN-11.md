---
code: MAIN-11
id: MDU6SXNzdWU4MjM2ODc5MjU=
databaseId: 823687925
number: 11
url: https://github.com/octomation/maintainer/issues/11
title: "integrate markdown parser"
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
createdAt: 2021-03-06T16:28:34Z
updatedAt: 2026-11-21T05:33:44Z
lastEditedAt: 2026-11-20T15:42:24Z
closedAt: 2026-11-21T05:33:44Z
---

# integrate markdown parser

## Motivation

The README already holds facts about a project that maintainer could use: the
links to its mirrors, quality reports, documentation and other integrations.
Many of them are badges written as reference definitions, so a regular
expression over the text reads them badly. The idea was to process Markdown
structurally, to extract this project information and to validate the
documentation against required sections.

An example from the README of this repository:

```text
Input:      [![Mirror][mirror.icon]][mirror.page]
Definition: mirror.page → https://bitbucket.org/kamilsk/maintainer
Result:     the link to the repository mirror
```

The historical candidate was [blackfriday](https://github.com/russross/blackfriday);
the choice of a library is not the user-facing outcome of this task.

## PoC

```bash
$ maintainer docs links README.md
# mirror:   https://bitbucket.org/kamilsk/maintainer
# docs:     https://maintainer.octolab.org/
# coverage: https://app.codecov.io/gh/octomation/maintainer
$ maintainer docs check README.md
# README.md: missing section "Installation"
```

## Acceptance criteria

- [ ] Headings, links and code blocks are told apart; text inside a code block
  never becomes project metadata.
- [ ] `[name][ref]` links, including nested badge links, resolve through their
  reference definitions.
- [ ] A missing required section is reported with the file and the section.
- [ ] A broken reference is reported, not dropped silently.

## Out of scope

- Generating a table of contents: [[MAIN-2]].
- Embedding source files into Markdown: [[MAIN-5]].
- Rendering Markdown to HTML.

Related: [[MAIN-2]], [[MAIN-5]].

<!-- 2026-11-20T21:00Z https://github.com/octomation/maintainer/issues/11#issuecomment-7043758253
Maturity: partly built, in another place. A Markdown parser is integrated: goldmark (go.mod) renders the bodies and comments of mirrored issues in the hub (internal/service/hub/markdown.go, added in 44af83a). Nothing extracts links from a README or checks its sections, and there is no docs command.
Proposal for v0.5.0: not planned. The parser the mission needs already serves the hub; README link extraction and section checks are documentation tooling, like the table of contents of #2, closed as not planned. The facts a contributor needs about a repository come from GitHub (homepage, topics), not from badges.
The title is kept for history: the parser is integrated, the README work is not.
-->

<!-- 2026-11-21T05:30Z https://github.com/octomation/maintainer/issues/11#issuecomment-7057976511
Closed as not planned: this is no longer the focus of maintainer. From v0.5.0 maintainer is the contributor's central tool: one picture of the repositories, issue trackers and GitHub across all projects, not build or release tooling.
goldmark already renders issue bodies in the hub. README checks are recorded on #153 as a candidate for an external tool.
-->
