---
code: MAIN-154
id: null
databaseId: null
number: null
url: null
title: "skill: ship an agent skill for the maintainer CLI and publish it to octolab/skills"
issueType: Task
labels:
  - "type: feature"
  - "scope: code"
  - "scope: test"
  - "scope: docs"
  - "impact: medium"
  - "effort: medium"
assignees:
  - kamilsk
milestone: "[[MAIN-M3]]"
parent: null
issueFields: {}
state: CLOSED
stateReason: COMPLETED
duplicateOf: null
createdAt: null
updatedAt: null
lastEditedAt: null
closedAt: null
---

# skill: ship an agent skill for the maintainer CLI and publish it to octolab/skills

## Motivation

Coding agents such as Claude Code and Codex already run `maintainer fetch` and `maintainer status` for the user, but they learn the CLI from `--help` alone: they apply a plan nobody reviewed, read exit `0` of a plan with conflicts as success, guess config keys that the strict config rejects, or put a token on the command line. indexit solved the same problem with an agent skill that ships inside the binary and is published to the [octolab/skills](https://github.com/octolab/skills) catalog on every stable release. maintainer gets the same: one skill, `maintainer`, about the CLI and the workspace of checkouts it manages. It is not the backlog mirror skill `projects:maintainer`, which keeps GitHub issues in `.github/`; the name is the binary's, as the catalog requires, and the description says what the skill is for.

## PoC

```bash
$ maintainer skill info
maintainer skill:
  version     : 0.2.1
  range       : >=0.2.1 <0.3.0
  ...
$ maintainer skill install --agent claude-code,codex
$ maintainer skill status --agent claude-code,codex --exact && echo ok
ok
$ go test ./skills && go generate ./skills && git diff --exit-code skills
```

## Expected behavior

- `skills/maintainer/` holds the skill: `SKILL.md` within the Agent Skills spec and the OctoLab metadata contract (`name`, `metadata.tool` and the directory are `maintainer`, `metadata.version` is the release without `v`, `tool-version-range` follows the catalog policy), a generated command reference and short recipes. It teaches the agent to check `maintainer version` against the range first; to plan with `fetch` before `fetch --apply` and to apply only with the user's go-ahead; to check `summary.conflict` and the exit codes `0`–`3` of fetch and `0`–`2` of status; to find and validate the config (`fetch config validate`, the CUE schema, unknown keys exit `2`); to read state and pins without editing state by hand; to resolve tokens through profiles and never print them; to use `status` offline. Every command it mentions exists, and every flag it uses is one `--help` shows.
- `go test ./skills` fails when the reference is stale, an example in the skill uses an unknown flag, a working command goes unmentioned, the compatibility ranges disagree, the embedded copy differs from the source, or the frontmatter breaks the catalog contract; `go generate ./skills` rewrites the reference.
- The binary embeds the skill: `maintainer skill show`, `info`, `export`, `install` and `status` work offline, read no `.env` and need no token, as indexit's do. `install` writes the skill for Claude Code or Codex with an ownership marker and replaces only a copy that maintainer installed and nobody changed since; `status` checks the copies the agents load.
- Release archives and `go install` builds carry the skill. The release check requires `metadata.version` of the tagged commit to equal the tag; `cd` checks that the source, the archives and `maintainer skill info` agree on the digest and, for a stable tag GitHub verified, publishes the skill with `octomation/skills-action` and a catalog token minted from the OctoLab Releaser App; a prerelease publishes nothing to the catalog. `doctor` checks that the catalog token can be minted.
- A guide page describes installing the skill from the binary or the catalog, keeping it in step after an upgrade, and what the skill does not do. AGENTS.md and the workflows guide describe the skill in the release.

## Acceptance criteria

- [ ] `go test -race ./skills/... ./internal/skill/... ./internal/command/...` passes and `go generate ./skills` leaves no diff.
- [ ] `maintainer skill install --agent claude-code,codex` into an empty home, then `maintainer skill status --agent claude-code,codex --exact` exits `0`; `ci` checks it on every change.
- [ ] `node .github/scripts/release.mjs check vX.Y.Z` fails when `skills/maintainer/SKILL.md` names another version; `node --test .github/scripts/*.test.mjs` covers it.
- [ ] `make go-dist-check` builds archives that hold `skills/maintainer/`, and `node .github/scripts/release.mjs skill` accepts them.
- [ ] `cd.yml` has a `skill` job for stable tags only; `actionlint` passes.
- [ ] `catalog.json` of `octolab/skills` registers `maintainer` for `octomation/maintainer`, `skills/maintainer`, before the first release with the skill.

## Out of scope

The backlog mirror skill `projects:maintainer`; skills for other agents than Claude Code and Codex; publishing a prerelease to the catalog; the published release notes.
