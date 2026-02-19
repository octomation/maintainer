# Code mode

Question: can the maintainer release this code, and what must he fix first?

## Target

- Default: the changes from the last stable tag to `HEAD`, with the working tree.
- Without a tag, the default is `full`.
- Other targets: `A..B`, a path, `full`.

```sh
base="$(git describe --tags --abbrev=0 --exclude='*-rc*')"
git diff --name-only "$base" -- . ':!.github/issues' ':!.github/roadmap' ':!.github/reviews'
git ls-files --others --exclude-standard
```

- M is the number of files in the target. The second command lists the changed tracked files; add the new untracked files.
- For a path target, M counts each tracked file under the path, test files included, without generated files.
- Leave out the issue mirror, `.github/reviews/`, and the folders that the project marks as not product, such as docs and tools.
- Leave out generated files and test data: build output, mocks, lock files, fixtures under `testdata/`.
- A file under `testdata/` that a test command of the project runs is a test, not a fixture. It is in M, and the Tests lens reads it.
- Leave out documents (`*.md`, `LICENSE`). The Contract lens still reads the documents.
- Add each left-out folder to the pathspec of the second command, such as `':!docs'`.
- The pass size rule of `references/report.md` applies.
- Above the pass size, read in this order: code that writes, moves or deletes user files and state; its tests; the rest.

## Gates

The gates are the checks of the project. See `SKILL.md`, Project.

- Take the lint and test commands that the project names for a change: `AGENTS.md`, `CONTRIBUTING.md`, the `Makefile`, the CI workflows.
- Without them, run the standard commands of each language of the target.
- Examples of defaults: `go vet ./...` and `go test -race ./...`; `npm test`; `cargo clippy` and `cargo test`; `pytest`.
- Run only the gates that apply to the target. A part of the target without a gate gives `partial`.
- Do not run a command that writes to tracked files: formatters, code generators, `npm ci` in the tree.
- If a gate needs a tool or a dependency that is absent, set Gates to `partial`. Do not install it in the tree.

## Lenses

| # | Lens | What it checks | Where to look |
| --- | --- | --- | --- |
| 1 | Gates | The gates above. | The checks of the project |
| 2 | Correctness | Behavior against the issue's acceptance criteria and the guides; edge cases; error paths; concurrency; timeouts and cancellation; unbounded growth of requests or memory. | The changed code; the criteria of its issues |
| 3 | Data safety | Each write, move and delete of user files and state: atomic change, partial failure, repeat run. | The code that writes, moves or deletes files, and the code that saves state |
| 4 | Contract | Flags, exit codes, output formats, configuration schema, state compatibility. `--help` and the guides agree with the behavior. For a user interface: the design rules of the project. | Command definitions, configuration schemas, the guides of the project |
| 5 | Tests | A test fails when the behavior breaks. Paths of the Data safety lens have tests. | The tests of the changed packages, and the tests of the target under `testdata/` |
| 6 | Simplicity | Dead code, duplication, needless abstraction, complexity that hides a defect. | The files of the target |

Lens order for sorting: Gates, Correctness, Data safety, Contract, Tests, Simplicity.

## Reproduce

- Build the program outside the tree, into the scratch folder.
- Run it with scratch `HOME`, `XDG_CONFIG_HOME` and `XDG_STATE_HOME`, and without tokens.
- Run it with the scratch folder as the working directory, so that no command writes into the tree.
- Compare the `--help` output of a command with its guide.
- For a scratch test in the tree, use an overlay of the toolchain, such as `go test -overlay <overlay.json>`. Keep the overlay files in the scratch folder.
- Without an overlay, copy the package to the scratch folder and test the copy.
- To test the Tests lens, break the line under test in the overlay or the copy. A good test then fails.
- Read the test before you mutate the line under test.
- Never edit generated files or build output in the tree.

## Severity examples

| Severity | Example |
| --- | --- |
| `Critical` | The default path of a command deletes a user file. |
| `High` | The default path prints a wrong result, or a rare state corrupts the state file. |
| `Medium` | A wrong exit code on a usage error; a page layout that breaks on a narrow screen. |
| `Low` | Dead code; a name that misleads; a write path without a test; a difference from the design rules that a user cannot see. |

## Not in this mode

- Attackers and their paths: security mode.
- The truth of the backlog: issue mode.
- The wording of the guides beyond the Contract lens; development tools; CI speed.
