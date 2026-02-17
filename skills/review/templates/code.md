---
uid: <uid>
timestamp: <timestamp>
model: <model>
---

# Code review

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
| Reviewed | <N> of <M> files |
| Excluded | <excluded or none> |
| Accepted | <accepted or none> |

## Findings

| ID | Severity | Lens | Location | Finding | Tracked |
| --- | --- | --- | --- | --- | --- |
| F1 | <severity> | <lens> | `<path>:<line>` | <finding> | <tracked> |

### F1. <finding>

- **Failure:** <condition, then the wrong result>
- **Evidence:** <reproduced or inspected> — <command and result, or path:line>
- **Fix:** <what must become true>

## Coverage

| Lens | Checked | Findings | Evidence |
| --- | --- | --- | --- |
| Gates | <checked> | <n> | <evidence> |
| Correctness | <checked> | <n> | <evidence> |
| Data safety | <checked> | <n> | <evidence> |
| Contract | <checked> | <n> | <evidence> |
| Tests | <checked> | <n> | <evidence> |
| Simplicity | <checked> | <n> | <evidence> |
