# Language

Every report, and every file of this skill, obeys these rules. `review.py check` tests the sentence limit and the banned words.

## From ASD-STE100, Issue 9

The writing rules apply. The dictionary is not enforced.

- A descriptive sentence has at most 25 words. An instruction has at most 20 words.
- Write one instruction in one sentence. Put the condition before the instruction.
- Use the active voice in instructions.
- STE permits the passive voice in a description when the agent is unknown or not important.
- Use simple tenses: the present for facts, the past for what you ran.
- Give one word one meaning. Give one thing one name.
- A noun cluster has at most three words.
- A paragraph has at most 6 sentences. Write a complex statement as a vertical list.
- An identifier, a command, a path or a quoted text in backticks counts as one word.

## House rules

- A report uses one limit: 25 words for each sentence.
- Use the active voice in each sentence of a report.
- Write `finding`, not `problem`, `defect` and `issue` in turn. `issue` means a backlog issue only.
- Put technical names in backticks.
- Write no home path that holds an account name. A generic path such as `~/.config/git` is valid.
- Write no names of private hosts, no local account names and no links to private notes.
- Public names such as `app.localhost` and `github.com` are valid.
- Do not use these words in authored prose, outside backticks and outside quotes:

`seems`, `appears`, `might`, `may`, `could`, `possibly`, `probably`, `perhaps`, `very`, `really`, `quite`, `extremely`, `comprehensive`, `robust`, `crucial`, `overall`, `basically`, `simply`, `note that`, `in order to`, `it is worth`.

The check matches a whole word without regard to case. The month `May` with a capital letter passes.

## Word count

- The check counts a backtick span as one word.
- A backtick span has at most 90 characters.
- A wiki link such as `[[ACME-131]]` counts as one word.
- A token without a letter or a digit, such as `—` or `|`, does not count.

## Before and after

The "before" lines come from real reports of an example project, shortened.

1. Scope as prose.
   - Before: "Reviewed the whole Go codebase of `acme/tool` at commit `2850b17 2026-07-07`: module manifests, CI/CD workflows, `Makefile`, …"
   - After: the `Scope` rows `` `2850b17` on `main`, clean tree `` and `N of M files`.
2. A legend of private keys.
   - Before: "Severity: **Crit** / **High** / **Med** / **Low** / **Nit**. Class: `correctness` · `security` · `concurrency` …"
   - After: no legend. The severity is `Critical`, `High`, `Medium` or `Low`. The `Lens` column names the cause.
3. A later annotation.
   - Before: "Revalidated on 2026-09-26 at `932b9c8`. See current statuses; the original findings below are historical."
   - After: nothing. The report is frozen. The next review is a new report, and the backlog holds the status.
4. A vocabulary made for one report.
   - Before: "“Relevant” means work remains; this review does not assess priority. “Partial” means some of the expected behavior is implemented."
   - After: the fixed keys of `references/report.md` only.
5. Evidence as a general claim.
   - Before: "Almost every finding was reproduced. Each finding says whether it was reproduced or judged by reading."
   - After: "**Evidence:** reproduced — a scratch test leaves a 0-byte `state.json`."
6. A failure told as code.
   - Before: "`Sync.Clone` calls `git.PlainCloneContext` at the requested target and, on **every** returned error, immediately runs `os.RemoveAll(opt.Path)`."
   - After: "**Failure:** A clone into an existing folder fails. `Sync.Clone` then deletes the folder and the user files in it."
7. A summary without a verdict.
   - Before: "New findings: **3 defect confirmed** — one High/P1 and two Medium/P2. Existing code-review entries: **10 no longer relevant, 4 defect confirmed** …"
   - After: "**BLOCKED** — F1 writes through a symlink parent outside the workspace." The counts table follows.
8. A hedge without a failure.
   - Before: "This validates that specific calendar; it does not guarantee resilience to all HTML changes."
   - After: nothing. A statement without a named failure is not a finding.
