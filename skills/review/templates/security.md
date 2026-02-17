---
uid: <uid>
timestamp: <timestamp>
model: <model>
---

# Security review

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

- **Failure:** A<n>: <condition, then the wrong result>
- **Evidence:** <reproduced or inspected> — <command and result, or path:line>
- **Fix:** <what must become true>

## Coverage

| Lens | Checked | Findings | Evidence |
| --- | --- | --- | --- |
| Gates | <checked> | <n> | <evidence> |
| Files | <checked> | <n> | <evidence> |
| Processes | <checked> | <n> | <evidence> |
| Secrets | <checked> | <n> | <evidence> |
| Service | <checked> | <n> | <evidence> |
| Supply chain | <checked> | <n> | <evidence> |
