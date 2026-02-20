# Report form

One form serves the three modes.

`review.py check` tests the form: the sections, the vocabularies, the counts, the order, the budgets, the links and the location of the file. It does not test if a finding is true, if the severity is right, or if the evidence is sufficient. Those are the reviewer's.

`review.py status` and `check --published` test the publication rule. See Publication.

`check` uses no network. Without a mirror, `status` and `check --published` read GitHub with `gh`.

## File

| Mode | Location |
| --- | --- |
| `code`, `issue` | `.github/reports/<UTC>Z.md` |
| `security` | `.agents/reports/<UTC>Z.md` |

- A security report is in `.agents/reports/` until its publication. A code or issue report is in `.github/reports/` or, when private, in `.agents/reports/`.
- `<UTC>` is `date -u +%Y%m%dT%H%M%S`. `review.py new` makes the name.
- Git must ignore `.agents/reports/`. A security report is private from its first line.
- `new security` refuses to write outside a Git repository or to a folder that Git does not ignore.
- A code or issue finding that an attacker can use makes the report private. Move the report to `.agents/reports/` before you write that finding.
- A private report moves to `.github/reports/` only by the rule of Publication.

## Front matter

Exactly three keys:

```yaml
---
uid: 5584fbb3-4e94-4399-997d-fc65767a7d91
timestamp: 2026-11-25T10:15:00Z
model: claude-opus-5.5-medium
---
```

- `uid`: a uuid4.
- `timestamp`: the same instant as the file name, ISO 8601 in UTC.
- `model`: the provider's model id with the real reasoning level as a suffix.

## Skeleton

- The H1 names the mode: `# Code review`, `# Issue review` or `# Security review`.
- Then exactly four H2 sections in this order: `## Verdict`, `## Scope`, `## Findings`, `## Coverage`. No other H2.
- The only H3 is a finding block. No H4.
- No fenced code block, no HTML, no image, no horizontal rule.
- Each section holds only its form. `check` rejects any other line:

| Section | Form |
| --- | --- |
| `## Verdict` | The verdict line and the counts table |
| `## Scope` | The scope table |
| `## Findings` | `None.`, or the findings table and the blocks |
| `## Coverage` | The coverage table |

No line stands between the H1 and `## Verdict`.

A complete sample report of an example project with the issue prefix `ACME`:

```markdown
# Code review

## Verdict

**BLOCKED** — F1 loses the state file when the disk is full.

| Critical | High | Medium | Low |
| --- | --- | --- | --- |
| 1 | 0 | 1 | 1 |

## Scope

| Field | Value |
| --- | --- |
| Revision | `1848546` on `main`, clean tree |
| Target | `v0.5.0..HEAD` |
| Reviewed | 14 of 14 files |
| Excluded | none |
| Accepted | [[ACME-93]] |

## Findings

| ID | Severity | Lens | Location | Finding | Tracked |
| --- | --- | --- | --- | --- | --- |
| F1 | Critical | Data safety | `internal/state/store.go:117` | A failed save leaves an empty state file | — |
| F2 | Medium | Contract | `internal/command/sync/root.go:104` | A negative `--concurrency` value is accepted | [[ACME-131]] |
| F3 | Low | Simplicity | `internal/view/view.go:125` | The header counter repeats the row filter | — |

### F1. A failed save leaves an empty state file

- **Failure:** The disk fills during `sync --apply`. `Store.Save` truncates the state file before it writes. The next run sees no records and plans a clone for each checkout.
- **Evidence:** reproduced — a scratch test with a 1 KB file system limit leaves a 0-byte `state.json`.
- **Fix:** A save must replace the state file only after the new content is complete on disk.

### F3. The header counter repeats the row filter

- **Failure:** The maintainer must change the same predicate in two places for each new filter.
- **Evidence:** inspected — `view.go:125` and `view.go:160` hold the same predicate.
- **Fix:** One function must hold the predicate.

## Coverage

| Lens | Checked | Findings | Evidence |
| --- | --- | --- | --- |
| Gates | full | 0 | `make check` exits 0 |
| Correctness | full | 0 | Read the 14 files of the target |
| Data safety | full | 1 | Ran 3 scratch tests on the save path |
| Contract | full | 1 | Compared `--help` output with `docs/sync.md` |
| Tests | full | 0 | Read the tests of the changed packages |
| Simplicity | full | 1 | Read the 14 files of the target |
```

## Verdict

Line 1 is `**<VERDICT>** — <reason>.` The reason is one sentence of at most 25 words. The reason does not restate the counts.
Derive the verdict from the tables. Take the first row that applies:

| Verdict | Condition |
| --- | --- |
| `BLOCKED` | At least one `Critical` or `High` finding. |
| `INCOMPLETE` | No `Critical` or `High` finding, and at least one lens is not `full`. |
| `ACTION` | Each lens is `full`, and the highest finding is `Medium`. |
| `CLEAR` | Each lens is `full`, and no finding is above `Low`. |

- A `BLOCKED` reason names each blocking finding by ID. Above three, name the first three and add `and N more`.
- An `INCOMPLETE` reason names the lenses that are not `full`.
- The counts table has one row. It counts all rows of the findings table, tracked rows included.

## Scope

Five rows, in this order, in each mode. No value is empty.

| Field | Value |
| --- | --- |
| Revision | `` `<short sha>` on `<branch>`, clean tree `` or `` `<short sha>` on `<branch>`, dirty tree: N files `` |
| Target | What the mode reviewed: a range, a path, `full`, or an issue set. |
| Reviewed | `N of M files` (code, security) or `N of M issues` (issue). |
| Excluded | What the request leaves out of the target, or `none`. |
| Accepted | The accepted decisions that already hold a candidate, or `none`. |

- `Revision`: N in `dirty tree: N files` is the line count of `git status --porcelain -- . ':!.github/reports' | wc -l`.
- `Reviewed`: N counts the items that you read in full. A search hit does not count.
- If N < M, at least one lens is not `full`.
- `Excluded` names only what the request leaves out. A default target has `none`.
- Do not list the exclusions of the mode reference, such as docs, tools, documents and fixtures. They apply to each target.
- `Accepted`: name the issue or the document path of each decision. Rule 5 of the admission bar defines a decision.

## Pass size

- The pass size is a threshold: 6 000 lines of target files (code, security) or 60 issues.
- At step 2, count the lines of the target list of the mode with `wc -l`.
- The target list holds the tracked files and the new untracked files.
- Read a target within the pass size in full.
- Then `partial` and `none` need a cause outside you: a missing tool, a missing token or no network.
- Above the pass size, read in risk order. Finish each file that you start.
- Stop after the file with which your read lines reach the pass size.
- `Reviewed` counts the files read in full. N < M follows from the files not read.
- The `Evidence` of each lens that is not `full` names what is not read and the cause.
- Lack of time is not a cause.
- A target that the request names stays the target above the pass size. Do not narrow it.
- Above the pass size, one request gives one pass and one report. Do not stop before the pass size.
- Split the target into passes only on an order of a full review.

### Full review in passes

- A target above the pass size gets a full review as several passes. The owner orders it.
- Split the target into parts, each within the pass size. One pass reviews one part and writes one pass report.
- A pass report is a normal report. Its `Target` is the part.
- A pass report lives in `.agents/reports/`, in each mode.
- The merge writes one report from the pass reports.
- The merge joins the findings of one cause, sorts and numbers them, and sums `Reviewed`.
- In the merged report, each lens takes the weakest `Checked` value of its passes. Its `Evidence` names the passes.
- After the merged report passes `check`, delete the pass reports.

## Findings

With zero findings, the section body is the one line `None.` and nothing else. With findings, `None.` is absent.

Otherwise, write a table with the columns `ID`, `Severity`, `Lens`, `Location`, `Finding`, `Tracked`.

| Column | Rule |
| --- | --- |
| `ID` | `F1`…`Fn`, in table order, without gaps. |
| `Severity` | `Critical`, `High`, `Medium` or `Low`. |
| `Lens` | The one lens of the mode that found the cause. |
| `Location` | Code and security: `` `path:line` ``, relative to the repository. Issue: an issue link; with a mirror, also a milestone or board file of the mirror, such as `[[ACME-M4]]`. |
| `Finding` | The defect as a statement, at most 15 words, no period at the end. It is not the fix. |
| `Tracked` | `—`, or the open issue that already holds this finding, as an issue link. |

- Sort by severity (`Critical`, `High`, `Medium`, `Low`), then by the lens order of the mode, then by `Location` as text.
- Give at most three locations, separated by `, `. Then add ` +N more`.
- An issue link has the form of the project. With a mirror, it is a wiki link with the prefix of the mirror, such as `[[ACME-131]]`.
- Without a mirror, an issue link is `#131`, or `owner/repo#131` for an issue of another repository.
- `review.py` trusts the state of an issue of another repository. Its owner can close it and make a private report publishable.
- With a mirror, a `Tracked` issue file must exist in `.github/issues/`, and the issue must be open.
- Without a mirror, `check` does not read GitHub. Check that the `Tracked` issue is open with `gh issue view`.
- The `Tracked` cell holds only an issue that was open at the review. A later issue links the finding from its `Source:` line. See Publication.
- One cause is one finding. Merge the symptoms of one cause.

## Finding blocks

Each row with `Tracked` `—` has one block, in table order. A tracked row has no block: the issue holds the details.

```markdown
### F1. <the Finding cell, word for word>

- **Failure:** <the scenario: condition, then the wrong result>
- **Evidence:** <reproduced — … | inspected — …>
- **Fix:** <what must become true>
```

| Field | Rule |
| --- | --- |
| `Failure` | A concrete scenario. Security: it starts with the attacker code, `A1:` to `A4:`. |
| `Evidence` | Starts with `reproduced — ` or `inspected — `. Cites a command with its result, or `path:line`. |
| `Fix` | The state that must become true. Name the outcome. Do not name the mechanism, the tool or the interface. Issue: it starts with the fix verb. |

- Write exactly these three fields, in this order. No field is empty.
- `reproduced`: you ran it and saw the failure.
- `inspected`: the cited lines, documents, tests or API results show the failure.
- Each claim of `Failure` has support in `Evidence`: a result that you saw, or a cited `path:line`.
- Do not state a result that you did not run as if you ran it.
- In `Evidence`, write the decisive setup and the command as you ran it.
- `Evidence` holds at least one command or `path:line` in backticks.
- `Evidence` can hold both kinds: first `reproduced — …`, then `inspected — …`.
- A backtick span has at most 90 characters. `check` rejects a longer span.
- Shorten a long command to its decisive part.
- When you shorten a command, keep each decisive input: the method and the path, each header the server checks, the body, the precondition.
- Put each decisive input in its own short span.
- Issue fix verbs: `Close`, `Rewrite`, `Split`, `Merge`, `Reopen`, `Untick`, `Move`, `Sync`.
- Budget: each field has at most 3 sentences and 50 words.
- In a `Low` finding, each field has 1 sentence of at most 25 words.
- In a `Low` finding, `Failure` names the cost that the maintainer pays now. It does not name a failure that a future change can cause.

## Coverage

Columns: `Lens`, `Checked`, `Findings`, `Evidence`. The rows are the six lenses of the mode, in the order of the mode.

| Column | Rule |
| --- | --- |
| `Checked` | `full` (the lens covered each item of the target that it applies to), `partial` (a part of them) or `none`. |
| `Findings` | The number of findings table rows with this lens. |
| `Evidence` | What you ran or read, at most 25 words. For `partial` and `none`, add the reason and what is not checked. |

A missing tool, token or network gives `partial` or `none`. It is not a finding.

A lens with no item in the target is `full`. Its `Evidence` says that no item applies.

## Vocabularies

The keys are fixed. No free text joins a key. There is no other vocabulary: no `Nit`, no `Info`, no class, no confidence, no status, no size.

Set the severity by impact first, then by reach:

| Impact | Default path | Needs a precondition |
| --- | --- | --- |
| Irreversible loss or compromise: user data, credentials, the release channel | `Critical` | `High` |
| Recoverable failure: a documented behavior is wrong | `High` | `Medium` |
| No failure for the user; the maintainer pays | `Low` | `Low` |

Disclosure of private data is irreversible.

A precondition is an uncommon configuration, a rare state, or an action that the attacker needs from the user or from the local machine.

Each severity demands one action:

| Severity | Action |
| --- | --- |
| `Critical` | Stop. Fix it now. |
| `High` | Fix it before the next release. |
| `Medium` | Plan it: make an issue. |
| `Low` | Fix it when you change this area. |

The mode references give examples. The examples obey the matrix; they do not override it.

## Admission bar

A candidate becomes a finding only if all of these are true:

1. Verified: `reproduced` or `inspected`. A suspicion without evidence is not a finding.
2. Observable: the finding names what fails, and for whom.
3. Actionable: the maintainer can make the change.
4. Not a copy of a tool: no configured linter or gate already reports it. A failed gate gives one finding for each independent cause.
5. Not accepted: no decision of the owner holds a rationale that still applies. A decision is an issue closed as not planned. It is also a limit that a document of the repository states as a decision: a guide, `AGENTS.md`, a README, or a settings file.

- If a decision applies, name its issue or document path in `Accepted` and drop the candidate.
- If a documented limit covers the case only in part, the rest is a finding.
- If the evidence or the exposure changed, the candidate is a finding.
- An open issue does not remove a finding. The finding stays in the table and in the counts. The `Tracked` cell names the issue.
- Before you write, search the issues of the project by path and by key term.

These are not findings: taste, hypothetical future problems, praise, generic hardening advice without a named failure.

## Publication

A private report becomes public only by this rule. `review.py` applies the rule to the report and to the issues of the project.

- With a mirror, `review.py` reads the files of `.github/issues/`.
- Without a mirror, it reads GitHub with `gh`: each issue of the repository with its comments, and each `Tracked` issue.
- If `gh` is absent, not signed in, or cannot read an issue, the rule does not hold. The report stays private.

The report does not change. It gets no status and no link to a later issue. The link goes from the issue to the report.

An issue for a finding of a private report holds this line. The line lists the ID of each finding that the issue fixes:

```markdown
Source: [<link text>](../reports/<report file name>), F1, F3.
```

- The line starts at the first column and ends with the period. The link text is free. The link target is `../reports/` and the file name of the report.
- An example of a target: `../reports/20261125T101500Z.md`. An example of a link text: `the security review of 2026-11-25`.
- On GitHub, a relative link does not resolve. There the target can be the URL of the report, under `.github/reports/` of a branch.
- An example of a URL target: `https://github.com/acme/tool/blob/main/.github/reports/20261125T101500Z.md`.
- The line names a finding by its ID. It does not quote the finding.
- Only an issue of a private report uses this form. The rule applies to each report that an issue names in this form.
- The link resolves after the publication.
- An issue names a report when its file holds the file name of the report without `.md`. Without a mirror, its body or a comment holds it.
- Each issue that names a private report must hold the `Source:` line. Each ID of the line must be in the report.

| Term | Meaning |
| --- | --- |
| The issues of a finding | The issue of its `Tracked` cell, and each issue whose `Source:` line lists its ID. |
| Closed as completed | The issue file has `state: CLOSED` and `stateReason: COMPLETED`. Each key is on one line of this exact form and occurs one time; a key in quotes counts. Without a mirror, `gh` gives the same two values. |
| A closed finding | The finding has at least one issue, and each of its issues is closed as completed. |

The rule: each finding of the report is closed, and each issue that names the report is closed as completed.

- A finding without an issue keeps the report private. This applies also to a finding that the owner fixed at once.
- An issue closed as not planned keeps the report private.
- A report without findings obeys the rule.

`status` gives each report one state:

| State | Folder | Meaning |
| --- | --- | --- |
| `private` | `.agents/reports/` | The rule does not hold. `status` prints each blocker. |
| `publishable` | `.agents/reports/` | The rule holds. `status` prints the `Publish:` line. |
| `published` | `.github/reports/` | The rule holds. |
| `refused` | `.github/reports/` | The rule does not hold, or `check` fails. `status` exits 1. |
| `duplicate` | `.agents/reports/` | A private copy of a published report. Delete it. |

```text
.agents/reports/20261201T090000Z.md: private, 1 of 2 findings closed.
.github/issues/ACME-190.md:22: publication: ACME-190 holds F2 and is not closed as completed
#191:1: publication: #191 holds F1 and is not closed as completed
.agents/reports/20261125T101500Z.md: publishable, 5 of 5 findings closed.
Publish: python3 <skill-dir>/scripts/review.py check --published .agents/reports/20261125T101500Z.md && mv .agents/reports/20261125T101500Z.md .github/reports/20261125T101500Z.md
```

- `status` lists each file of `.agents/reports/`.
- In `.github/reports/`, `status` lists each security report, each report with a private copy, and each report that an issue names by finding ID.
- The `Publish:` line tests the rule first. If the test fails, the file stays private.
- The `Publish:` line moves the file. The public file equals the private file byte for byte.
- `check --published` tests one report against the rule, in each of the two folders. A closed `Tracked` issue is valid there.
- `check` without `--published` is the check of the review: it demands an open `Tracked` issue.
- `check` without `--published` refuses a security report in `.github/reports/`.
- If the two folders hold a file with the same name, the two files must be equal. The `Publish:` line does not replace a different file.
- A blocker line has the form of a `check` error. It names the issue file or the report line to correct. Without a mirror, it names the issue as `#N`.
