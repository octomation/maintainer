---
name: review
description: >-
  Write a review report of a repository in a fixed form, in one of three
  modes, and check the form: code mode asks if the maintainer can release the
  code, issue mode asks if the backlog states the truth, security mode asks
  who can harm the user or the release channel. The report is one file in
  .github/reviews/, or in a private, Git-ignored .agents/reviews/ for
  security; it is not a set of comments on a diff. It also tells when a
  private report is ready for publication. Use on "review report", "write a
  code review report", "write an issue review report", "write a security
  review report", "publish the security report", "отчёт ревью", "опубликовать
  отчёт", or maintainer:review.
license: MIT
compatibility: >-
  Requires python3 and Git. Without an issue mirror in .github/issues/, the
  publication rule reads GitHub through a signed-in gh.
metadata:
  author: octomation
  version: "0.6.0"
  tool: maintainer
  tool-version-range: ">=0.6.0 <0.7.0"
---

# Review

## Frame

- You are the reviewer. The reader is the maintainer. He works alone. He reads a report in two minutes.
- A false finding costs more than a missed `Low` finding: it spends the only person.
- Write only what leads to an action.
- Do not write praise, a summary of the code, a story of the review or hedge words.
- Show exceptions. Count healthy items; do not describe them.
- Do the steps in order. Each step has one output. A lens that you did not apply in full shows in `Coverage`.

## Modes

| Mode | Question | Default target | Reference |
| --- | --- | --- | --- |
| `code` | Can the maintainer release this code, and what must he fix first? | The last stable tag to `HEAD`, with the working tree | `references/code.md` |
| `issue` | Can the maintainer plan work from the backlog, and where does the backlog state something false? | Each open issue, and each issue closed after the last stable tag | `references/issue.md` |
| `security` | Who can harm the user or the release channel, and how? | `full`: the product at `HEAD` | `references/security.md` |

Read the mode reference and `references/report.md` before step 2. Read `references/language.md` before step 8.

## Project

The reviewed project gives three inputs. Read them in step 2, from `AGENTS.md`, `CONTRIBUTING.md`, `README.md`, `SECURITY.md`, the `Makefile` or the CI workflows.

| Input | From the project | Default without it |
| --- | --- | --- |
| Checks | The lint and test commands that the project names for a change | The standard lint and test commands of each language of the target |
| Threat model | The attackers and the assets that the project names | The four attackers of `references/security.md` |
| Issue codes | The mirror: files `.github/issues/<PREFIX>-<n>.md` | GitHub issues: `#N` or `owner/repo#N` |

- A check that writes to tracked files is not a gate. Use its read-only form, or set Gates to `partial`.
- `review.py` takes the prefix from the file names of the mirror. A project without a mirror has no prefix.
- With a mirror, an issue link is a wiki link such as `[[ACME-131]]`. Without a mirror, it is `#131` or `owner/repo#131`.
- Name each default that you used in the `Evidence` of its lens.

## Procedure

Steps 1–10 are yours. Steps 11–13 are the owner's.

| # | Step | Action | Output |
| --- | --- | --- | --- |
| 1 | Request | Take the mode and the target. Without a target, use the default target of the mode. | Mode and target |
| 2 | Scope | Read the inputs of the project. Resolve the revision and the target. Count the items (M) and the lines. Record the tree digest. Run `new`. | Report file, M, lines, digest |
| 3 | Gates | Run the gates of the mode. Gates must not write to tracked files. | Gate results |
| 4 | Hunt | Apply lenses 2–6 of the mode in order. | Candidates |
| 5 | Verify | Reproduce or inspect each candidate. Drop the rest. | Verified candidates |
| 6 | Match | Search the issues of the project by path and by key term for each candidate. | `Tracked` cells, `Accepted` codes |
| 7 | Rate | Set the severity from the matrix. Sort. Number. | Findings table |
| 8 | Write | Fill the template. Obey `references/language.md`. | Filled report |
| 9 | Check | Run `check`. Correct the report until the check passes. Compare the tree digest with step 2. | Exit 0, same digest |
| 10 | Hand over | Print the lines that `check` prints. The report is frozen. | Hand-over lines |
| 11 | Triage (owner) | For each finding: fix now, make an issue, or accept it. | Backlog changes |
| 12 | Publish (owner) | Private report: when each finding is closed, run the `Publish:` line and push. | Public report |
| 13 | Next review (owner) | Order a new snapshot. | Next report |

A target within the pass size is read in full. Lack of time is not a cause for `partial`. See `references/report.md`, Pass size.

A full review of a target above the pass size runs as passes and one merge. See `references/report.md`, Pass size.

Step 6: search the mirror with `grep`; without a mirror, search with `gh issue list --state all --search`.

Step 6: an open issue gives `Tracked`. A decision of the owner with a rationale that still applies gives `Accepted`. A decision is an issue closed as not planned, or a limit that an issue or a guide states as the owner's decision.

Step 11: to accept a finding, the owner closes an issue as not planned, with the rationale. The report does not change.

Step 11, private report: the owner fixes first or drafts a private security advisory. The details never go into a public issue.

Step 11, private report: each finding gets an issue, also a finding that the owner fixes at once. The issue holds the `Source:` line of `references/report.md`, Publication.

Step 12: you propose, the owner publishes. Print the `Publish:` line. Do not run it.

Step 13: a fixed finding is absent. A tracked finding stays visible until its issue closes.

Commands, from the repository root. `<skill-dir>` is the folder of this file.

```sh
git status --porcelain -- . ':!.github/reviews' | shasum
python3 <skill-dir>/scripts/review.py new <code|issue|security> --model <model>
python3 <skill-dir>/scripts/review.py check <report>
python3 <skill-dir>/scripts/review.py status
python3 <skill-dir>/scripts/review.py check --published <report>
```

- Tree digest: the first command. Run it in step 2 and in step 9. The two results must be equal.
- `<model>`: the provider's model id with the real reasoning level as a suffix, such as `claude-opus-5.5-medium`.
- `new` prints the path of the report. It refuses to overwrite a file.
- `check` prints one line for each error: `<file>:<line>: <rule>: <message>`.
- `check` uses no network and changes no file.
- `status` prints the state of each private report and of each published report that the publication rule applies to.
- With a mirror, `status` reads the reports and `.github/issues/`. It uses no network and changes no file.
- Without a mirror, `status` and `check --published` read the issues with `gh`.
- If `gh` is absent, not signed in or cannot read an issue, the report stays private.
- `status` exits 1 when a report in `.github/reviews/` breaks the publication rule or when it cannot read the issues.
- `check --published` applies the publication rule to one report. It refuses a report with a finding that is not closed.
- `check` without `--published` is the check of step 9. After the hand-over, test a private report only with `status` or `check --published`.

## Private reports

- `new security` writes to `.agents/reviews/`. A security report is private from its first line.
- `new security` refuses to write when Git does not ignore the new report in `.agents/reviews/`, or outside a Git repository.
- `check` gives a `privacy` error for a report in `.agents/reviews/` that Git does not ignore.
- If `new` refuses, tell the owner to add `.agents/reviews/` to `.gitignore` or `.git/info/exclude`. Do not add it yourself.
- A code or issue finding that an attacker can use makes the report private.
- Before you write that finding, move the report to `.agents/reviews/` with the same file name.
- A private report stays private until each of its findings is closed. See `references/report.md`, Publication.
- Never move a report to `.github/reviews/` yourself. Only the owner publishes: he runs the `Publish:` line and pushes.
- The `Publish:` line tests the report with `check --published`. Then it moves the file unchanged.
- On a request to publish a report, run `status` and print its lines.

## Boundaries

- The deliverable is one report file. The tree, git and GitHub are read-only. Do not commit.
- Run reproductions outside the tree: in a scratch folder, or through an overlay of the toolchain, such as `go test -overlay`.
- Run a built program with scratch `HOME`, `XDG_CONFIG_HOME` and `XDG_STATE_HOME`, and without tokens.
- Run a built program with the scratch folder as the working directory.
- Do not read the owner's real configuration, state or data.
- Leave old reports in `.github/reviews/` as they are. Do not restate their status.
- Do not add a status column or a later annotation to a report. The backlog is the only registry.
- `status` derives the state of a report from the backlog. The report holds no state and no link to a later issue.
- Do not make issues. Do not fix findings.

## Hand-over

On success, `check` prints these lines. Print them as your last message, without other text, unless the request asks for more.

```text
Code review: BLOCKED — Critical 1, High 0, Medium 1, Low 1.
Report: .github/reviews/20261125T101500Z.md
First: F1 A failed save leaves an empty state file.
Coverage: full.
```

- `First:` lists the `Critical` and `High` findings, at most three. Without them it is `First: none.`
- `Coverage:` is `full.` or `partial: <lenses that are not full>.`
- A private report adds a fifth line: `Private: do not publish this report before the fix ships.`
- With a mirror, if the rule of Publication already holds for the new private report, the fifth line is its `Publish:` line.
- With a mirror, `check` adds the `Publish:` line of each other private report that is publishable.
- With a mirror, `check` adds a `Refused:` line for each report in `.github/reviews/` that breaks the publication rule.
- Without a mirror, run `status` for these lines: `check` does not read GitHub.

```text
Publish: python3 <skill-dir>/scripts/review.py check --published .agents/reviews/20261125T101500Z.md && mv .agents/reviews/20261125T101500Z.md .github/reviews/20261125T101500Z.md
Refused: .github/reviews/20261125T101500Z.md breaks the publication rule; run status.
```
