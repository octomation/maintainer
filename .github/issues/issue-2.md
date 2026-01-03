---
id: 2
database_id: 777286024
node_id: MDU6SXNzdWU3NzcyODYwMjQ=
status: closed
title: "implement doctoc"
labels: []
url: https://github.com/octomation/maintainer/issues/2
created_at: 2021-01-01T13:33:11Z
updated_at: 2023-03-31T15:35:21Z
---

# implement doctoc

Automate the table of contents of Markdown documents so that adding or renaming a section does not force the maintainer to fix links by hand. The reference is [doctoc](https://github.com/thlorenz/doctoc), which builds a table of contents from the document headings.

The user-visible result for a README with "Installation" and "Usage" sections:

```markdown
- [Installation](#installation)
- [Usage](#usage)
```

The dedicated table-of-contents block is expected to be updated while the rest of the text is preserved. Regenerating it without changing the headings must not produce a diff, and the links must resolve to their sections when the document is viewed on GitHub.

**Current state:** the issue is closed, but neither a doctoc command nor an equivalent integration exists in the current tree. The closed status does not mean the CLI has this capability. The related open task for Markdown processing is [#11](issue-11.md).
