# Development tools

`tools/go.mod` is a separate module that declares the development tools with Go's `tool` directive. It follows the current stable Go release, as the application does.

```sh
make tools                  # install into bin/<os>/<arch>
source bin/activate         # put these tools on PATH
make tools-tidy             # maintain tools/go.mod and tools/go.sum
make tools-check            # verify modules and scan the tool packages
make tools-update           # go get tool, tidy, reinstall
```

| Tool | Used for |
| --- | --- |
| `cue` | `make config-vet`: validate `.github/settings.json` against `.github/settings.cue` |
| `golangci-lint` | `make lint`; the version matches `ci.yml` |
| `goreleaser` | `make dist-check`, `make release-config-check`; the version matches `cd.yml` |
| `govulncheck` | `make deps-check` for the application, `make tools-check` for the tools |
| `mockgen` | `make generate`: the mocks of the application |
| `mod` | `./Taskfile github <major>`: upgrade go-github to a new major version |
| `goimports` | `make format` |
| `gorelease` | the `release` helper of `bin/activate`: check API compatibility against the latest tag |
| `benchcmp`, `gomvpkg`, `gorename` | code maintenance by hand; `gorename` is deprecated upstream |

Every binary comes from `go install tool`, and `tools.yml` checks that `go list tool` matches the installed executables. Modules behind the `tool` directive are required as `// indirect`, so the tool pattern is the only way to select them.

Dependabot skips indirect requirements, so its `/tools/` entry raises only security updates ([dependabot-core#12050](https://github.com/dependabot/dependabot-core/issues/12050)). Update with `make tools-update` and review the diff; when `golangci-lint` or `goreleaser` moves, update the pinned versions in the workflows too.

`tools-check` runs only in `tools.yml`: a vulnerability in a tool fails that workflow, not the application CI or a release, which scan the application with `deps-check`. It currently reports [GO-2026-6225](https://pkg.go.dev/vuln/GO-2026-6225) and [GO-2026-5932](https://pkg.go.dev/vuln/GO-2026-5932), which have no fixed version; neither is suppressed.
