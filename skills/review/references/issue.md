# Issue mode

Question: can the maintainer plan work from the backlog, and where does the backlog state something false?

## Target

- Default: each open issue, and each issue closed after the last stable tag.
- The start is the tagger date of the tag; commit dates are not reliable. The owner can name another tag, such as "closed after v1.4.1".
- Without a tag, the default is each open issue.
- Other targets: a milestone, a board column, or a list of issues.

```sh
tag="$(git describe --tags --abbrev=0 --exclude='*-rc*')"
git for-each-ref --format='%(taggerdate:iso-strict)' "refs/tags/$tag"
```

With a mirror, `<PREFIX>` is the prefix of its file names:

```sh
grep -l '^state: OPEN' .github/issues/<PREFIX>-*.md
grep -H '^closedAt:' .github/issues/<PREFIX>-*.md
```

Without a mirror:

```sh
gh issue list --state open --limit 1000 --json number,title,labels,milestone
gh issue list --state closed --search "closed:>=<tagger date>" --limit 1000 --json number,title,stateReason,closedAt
```

- A closed issue is in the target if its `closedAt` value is later than the tagger date. Compare instants, not text.
- M is the number of issues in the target.
- Above the pass size of 60 issues, the order is: In progress, To do, issues closed in the target, then Backlog.
- The roadmap overview is not part of this mode.

## The backlog

| What | With a mirror | Without a mirror |
| --- | --- | --- |
| Issue | `.github/issues/<PREFIX>-N.md`: front matter with `state`, `stateReason`, `milestone`, `labels`, `closedAt` | `gh issue view N --json state,stateReason,milestone,labels,closedAt,body` |
| Comments | The comment blocks at the end of the issue file | `gh issue view N --comments` |
| Milestone criteria | The milestone file of the mirror, if the mirror holds one | The milestone description on GitHub |
| Board | The board file of the mirror and the rule under each column | The project board on GitHub, if the project uses one |
| Accepted | Issues with `stateReason: NOT_PLANNED` | Issues closed as not planned |

- An example of a mirror: `.github/issues/` and `.github/roadmap/`, as the `projects:maintainer` skill writes them.
- With a mirror, `Location` is a wiki link to a file of the mirror, such as `[[ACME-131]]` or `[[ACME-M4]]`.
- Without a mirror, `Location` is `#N` or `owner/repo#N`.
- Before a `Close` or `Rewrite` finding, read the comments of the issue and the criteria of its milestone.

## Gates

- With a mirror, the gates are the offline check and the GitHub comparison of the mirror tool, if the project names them.
- An example of a mirror gate: `maintainer.py lint` and `maintainer.py verify` of the `projects:maintainer` skill.
- An offline check is a pass when it exits 0. A comparison that lists differences gives one `Sync` finding, `Low`.
- Without a mirror tool, the Gates lens is `full`, and its `Evidence` says that no gate applies.
- Without `gh` or the network, set Gates to `partial`. In the report, write a command without the home path.

## Lenses

| # | Lens | What it checks | Usual fix verb |
| --- | --- | --- | --- |
| 1 | Gates | The gates above: the mirror and GitHub agree. | `Sync` |
| 2 | Valid | The need exists on the reviewed revision: not done, not obsolete. | `Close` |
| 3 | Ready | An issue in the board columns To do or In progress states the expected behavior, the limits and the acceptance criteria. Raw Backlog issues pass. | `Rewrite` |
| 4 | Single | One deliverable in one issue. No duplicate. No conflict with another issue. | `Split`, `Merge` |
| 5 | True | An issue closed as completed is complete in the code. Each ticked criterion is true, in open issues too. | `Reopen`, `Untick` |
| 6 | Placed | Milestone, board column and labels obey the rules of the board and of the milestone. | `Move` |

- Ready: the rule for a ready issue comes from the project, such as `AGENTS.md` or `CONTRIBUTING.md`. Without one, the rule of the table applies.
- Without a board, Ready applies to each open issue in a milestone. The issues without a milestone pass.
- Valid and True: check the claim with `git grep`, the guides of the project and the built program.
- Placed: compare the column of each card with `state` and `stateReason` of the issue. Without a board and milestones, Placed checks the labels only.
- The `Fix` field starts with one fix verb: `Close`, `Rewrite`, `Split`, `Merge`, `Reopen`, `Untick`, `Move` or `Sync`.
- Any verb of the list is valid in any lens. Use the verb that names the action.

## Severity examples

The user of the matrix is the one who relies on the backlog. Delivered work is the default path. Work that waits to start needs a precondition.

| Severity | Example |
| --- | --- |
| `Critical` | The backlog states that a defect that loses data is fixed, and it is not. |
| `High` | The backlog states something false about delivered work: an issue closed as completed is not complete; a ticked criterion is false. |
| `Medium` | An issue in To do or In progress cannot run as written: no criteria, a duplicate, a conflict. |
| `Low` | The same in Backlog; a wrong column or label; stale wording; the mirror and GitHub differ. |

## Not in this mode

- The quality of the code: code mode.
- Attackers: security mode.
- The roadmap overview and the priority of issues.
