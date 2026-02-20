# Security mode

Question: who can harm the user or the release channel, and how?

## Target

- Default: `full`, the product at `HEAD`. The attack surface does not divide by diff.
- M counts only the files that hold an entry point or a sink. Leave out tests, fixtures, styles, mocks and documents.
- These are the files that read remote data, write files, start processes, hold tokens or serve requests.
- Add the CI workflows and actions, the release configuration, the install scripts and the build files, such as the `Makefile`.
- Above the pass size, read in risk order: services that accept requests; files that start processes or hold tokens; release workflows; the rest.

```sh
git grep -lE 'exec\.Command|os\.(WriteFile|Rename|RemoveAll|MkdirAll|CreateTemp|Create|OpenFile)|net/http|_TOKEN' -- '*.go' ':!*_test.go'
git ls-files -- .github/workflows .github/actions Makefile '*.sh'
```

The first command is an example for Go; write its equivalent for each language of the target. It is a start. Add each file that it misses and that holds an entry point or a sink.

## Private location

- `review.py new security` writes to `.agents/reports/`, which Git must ignore. `new` refuses a folder that Git does not ignore.
- Never move the report under `.github/` yourself. Only the owner publishes it.
- The report stays private until each finding is closed. See `references/report.md`, Publication.
- An issue for a finding states the behavior that must become true. It does not state the attack.
- The issue names the finding by ID in its `Source:` line.
- Hand-over adds the line `Private: do not publish this report before the fix ships.`
- When each finding is closed, hand-over adds the `Publish:` line in its place.

## Attackers

Each `Failure` starts with one attacker code. The codes are fixed; the project gives their content.

- Read the threat model of the project in `SECURITY.md`, `AGENTS.md` or a threat model document.
- Map each attacker that the project names to the code of the table that has the same access.
- An attacker that the project excludes is outside the model. Name the document in `Accepted` if a candidate needs it.
- Without a threat model, the default model of the table applies. Name it in the `Evidence` of the Gates lens.

| Code | Attacker | Controls |
| --- | --- | --- |
| A1 | Author of remote content | The data that the product reads from the network: repositories, files, names, URLs, text, API responses. |
| A2 | Web page in the user's browser | Requests to a local service of the product on loopback. |
| A3 | Another OS user on the machine | Files in shared folders; local ports. |
| A4 | Supply chain | A dependency, a CI action, a pull request from a fork, the release channel. |

A process that already runs as the user is outside the model: it has the user's rights.

## Gates

- The dependency audit of each ecosystem of the project, such as `govulncheck ./...`, `npm audit --omit=dev`, `pip-audit` or `cargo audit`.
- `actionlint` if the project has GitHub workflows.
- The secret scan of tracked files:

```sh
git grep -nIE '(ghp|gho|ghs|ghu|ghr)_[A-Za-z0-9]{30,}|github_pat_[A-Za-z0-9_]{30,}|-----BEGIN [A-Z ]*PRIVATE KEY|AKIA[0-9A-Z]{16}'
git ls-files | grep -E '\.env$|\.pem$|\.key$|id_(rsa|ed25519)$'
```

- No output of the secret scan is a pass. Leave out build output with a pathspec, such as `':!dist'`.
- The project can pin its audit tools. Then use the pinned tool.
- The audits need the network. An absent tool or no network gives `partial`.

## Lenses

| # | Lens | What it checks | Where to look |
| --- | --- | --- | --- |
| 1 | Gates | The gates above. | The commands above |
| 2 | Files | Names from remote data, configuration and URLs that become paths: traversal, symlinks, overwrite, temporary files, file modes. | The code that turns names into paths, and the code that writes files |
| 3 | Processes | Calls of `git` and other programs: option injection; hostile repository content, such as hooks, filters and configuration. | The code that starts processes |
| 4 | Secrets | Tokens: load, storage, transit; logs, errors, URLs, process arguments, file modes. | The code that reads, stores or sends tokens |
| 5 | Service | A service of the product that accepts requests: bind address, `Host` and `Origin` checks. Requests that change state, output escaping, resource exhaustion, exposure of private data. | The listeners, the handlers and the pages that they serve |
| 6 | Supply chain | Pins of dependencies and actions, workflow permissions, secrets in CI, code from fork pull requests, signatures, checksums, the install script, the release channel. | The dependency manifests and lock files, the CI workflows, the release configuration, the install scripts |

- Name the weakness class with its usual name, such as path traversal or DNS rebinding. No CWE number is required.
- Reproduce against the built program as `references/code.md` says. Use scratch repositories and scratch servers as A1 content.
- A product without a service has no item for the Service lens. The lens is then `full`.

## Scratch service

- Start the service of the product from the scratch build, with the scratch folders of `references/code.md`.
- Take a free port. Never use the port of an instance that the owner runs.
- Test each host name and address that the service answers. If `curl` does not resolve a `.localhost` name, send its `Host` header to `127.0.0.1`.
- A request that changes state needs the headers that the service demands. Send each one in a test of access.
- Stop the service before you write the report.

## Severity examples

| Severity | Example |
| --- | --- |
| `Critical` | A1 or A2 gets a token or runs code during normal use. |
| `High` | The same after one plausible user action; or A3 reads a token, reads private data or overwrites a user file. |
| `Medium` | A2 or A3 makes a service stop or show wrong data, and a restart repairs it. |
| `Low` | No attacker gains and the maintainer pays; for example, a gate of this mode cannot run because a tool pin is stale. |

A missing defense without a present attack path is not a finding.

## Not in this mode

- Defects without an attacker: code mode.
- A process that runs as the user.
- Generic hardening advice without a named failure.
