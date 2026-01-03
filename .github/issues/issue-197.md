---
id: 197
database_id: 2092219781
node_id: I_kwDOE2M9Zc58tL2F
status: open
title: "insights: deps: fetch data from deps.dev"
labels: ["scope: code","type: feature","impact: high","effort: medium"]
url: https://github.com/octomation/maintainer/issues/197
created_at: 2024-01-20T19:51:39Z
updated_at: 2024-01-20T19:51:39Z
---

# insights: deps: fetch data from deps.dev

Bring deps.dev information into dependency analysis so the maintainer can assess a specific package and version: the available versions, the licences, the dependency tree and the published advisories. These are the insights the source makes available through its API.

A provisional interface:

```bash
maintainer insights deps go github.com/spf13/cobra@v1.10.2
# The package and version, the data source, the dependencies and the available risk signals.
```

The ecosystem, the package and the version have to be distinguished explicitly; absent data does not mean absent problems. The result must show the provenance of the information and be suitable for further processing, for example through a proposed JSON mode. This is information retrieval, not automatic dependency updating.

**At present** there is no `insights` group and no integration. The original references are [deps.dev](https://deps.dev/), its [documentation](https://docs.deps.dev/) and the [v3alpha API](https://docs.deps.dev/api/v3alpha/); since that alpha contract was the starting point, the currently supported API version has to be chosen before implementing. Related context: [raw dumps](<../notes/Issues/maintainer as raw.md>), [the project graph](<../notes/Issues/project graph.md>), [version delta research](<../notes/Issues/version delta research.md>).
