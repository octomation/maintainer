---
uid: <uid>
timestamp: <timestamp>
model: <model>
---

# Issue review

## Verdict

**<VERDICT>** — <reason>.

| Critical | High | Medium | Low |
| --- | --- | --- | --- |
| <n> | <n> | <n> | <n> |

## Scope

| Field | Value |
| --- | --- |
| Revision | `<short sha>` on `<branch>`, <clean tree or dirty tree: N files> |
| Target | <target> |
| Reviewed | <N> of <M> issues |
| Excluded | <excluded or none> |
| Accepted | <accepted or none> |

## Findings

| ID | Severity | Lens | Location | Finding | Tracked |
| --- | --- | --- | --- | --- | --- |
| F1 | <severity> | <lens> | <issue> | <finding> | <tracked> |

### F1. <finding>

- **Failure:** <condition, then the wrong result>
- **Evidence:** <reproduced or inspected> — <command and result, or path:line>
- **Fix:** <Verb> <what must become true>

## Coverage

| Lens | Checked | Findings | Evidence |
| --- | --- | --- | --- |
| Gates | <checked> | <n> | <evidence> |
| Valid | <checked> | <n> | <evidence> |
| Ready | <checked> | <n> | <evidence> |
| Single | <checked> | <n> | <evidence> |
| True | <checked> | <n> | <evidence> |
| Placed | <checked> | <n> | <evidence> |
