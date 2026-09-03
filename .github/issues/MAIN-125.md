---
code: MAIN-125
id: null
database_id: null
node_id: null
status: open
title: "github: contribution: lookup panics without an explicit week range"
labels: ["scope: code","scope: test","type: bug","severity: major","impact: medium","effort: easy"]
url: null
created_at: 2026-09-01T11:11:10Z
updated_at: 2026-09-01T11:11:10Z
---

# github: contribution: lookup panics without an explicit week range

**Motivation:** `contribution lookup` is unusable in every form the changelog
documents as the default one. The command fails before it reaches the network,
so no argument spelling recovers it except passing a week range by hand.

**Details**

`ParseDate` derives the `Half` flag from the `/weeks` suffix. When the suffix is
present the flag follows the sign of the parsed value, so a negative range stays
non-half. When the suffix is absent the flag is set unconditionally, without
consulting the sign of `defaultWeeks`:

```go
// internal/command/github/contribution/helper.go
if rawWeeks != "" {
    ...
    if weeks > 0 && rawWeeks[0] != '+' {
        opts.Half = true
    }
} else {
    opts.Half = true // ignores the sign of defaultWeeks
}
opts.Weeks = weeks
```

`lookup` passes `-1` as `defaultWeeks`, so an omitted suffix yields
`{Weeks: -1, Half: true}`. `LookupRange` forwards it to `xtime.GregorianWeeks`
and then to `xtime.RangeByWeeks`, whose first statement asserts the combination
is impossible:

```go
// internal/pkg/time/range.go:17
assert.True(func() bool { return !half || (half && weeks > 0) })
```

Nothing ever assigns `assert.disabled`, so the check always runs: the process
panics with `assertion is not a true` and prints a stack trace instead of a
table.

`suggest` is unaffected only because it passes `5`, where the unconditional
`Half = true` happens to be the intended value.

Introduced in `40b655c` (2022-09-17, `issue #44: github: contribution: refactor
command and view, first step`), which added both `defaultWeeks` and the
`lookup` call site in the same change.

The regression survived because every case of `TestParseDate` fixes
`fWeeks: 5`; the `-1` default of `lookup` is never exercised, and no test calls
the command end to end.

Required behaviour:

- an omitted `/weeks` suffix resolves the `Half` flag from the sign of
  `defaultWeeks`, the same rule an explicit suffix follows;
- `lookup` without a suffix behaves as `<date>/-1`, the form the changelog
  documents, and renders the range that ends on the requested date;
- `suggest` keeps its current output for every documented invocation;
- `ParseDate` is covered for a negative `defaultWeeks`, both with and without a
  suffix, so the two branches can no longer diverge;
- no reachable input makes a command panic on an assertion.

**PoC**

Every invocation without an explicit `/weeks` suffix fails:

```bash
$ maintainer github contribution lookup
recovered: assertion is not a true
---
unexpected panic occurred
go.octolab.org/toolset/maintainer/internal/pkg/assert.True
	internal/pkg/assert/assert.go:33
go.octolab.org/toolset/maintainer/internal/pkg/time.RangeByWeeks
	internal/pkg/time/range.go:17
go.octolab.org/toolset/maintainer/internal/pkg/time.GregorianWeeks
	internal/pkg/time/range.go:50
go.octolab.org/toolset/maintainer/internal/model/github/contribution.LookupRange
	internal/model/github/contribution/helper.go:21
go.octolab.org/toolset/maintainer/internal/command/github/contribution.Lookup.func1
	internal/command/github/contribution/lookup.go:25
$ echo $?
1

$ maintainer github contribution lookup 2013-12-03 # same panic
$ maintainer github contribution lookup now        # same panic
$ maintainer github contribution lookup git        # same panic
```

Passing a range explicitly is the only workaround:

```bash
$ maintainer github contribution lookup /-1        # works
$ maintainer github contribution lookup 2013-12-03/9
$ maintainer github contribution lookup 2013/-10
```

The regression test should drive `ParseDate` with `defaultWeeks: -1` and assert
the resulting range, so the assertion is never reached.

<!--
## Context

Found while verifying [[MAIN-126]]: the plan was to run `lookup` a few times in
a row and compare the counts, and the command never got as far as the network.

The panic is not data-dependent. `LookupRange` is called before the service is
touched, so every invocation without a `/weeks` suffix fails identically,
whatever the profile or the token. That also means the fix cannot be validated
against GitHub — the whole failure lives between `ParseDate` and `RangeByWeeks`.

Both halves of the defect landed in the same commit, so `lookup` has never
worked in its documented default form since 2022-09-17. Only `suggest` exercised
`ParseDate` with a positive `defaultWeeks`, and the tests copied that one value
into all 18 cases, so the branch `lookup` depends on was never executed by
anything but the user.

The shape of the fix is one expression: the `else` branch should derive the flag
the same way the parsed branch does, from the sign of the value it is about to
store, i.e. `opts.Half = defaultWeeks > 0`. Worth checking `GregorianWeeks` for
negative `weeks` at the same time — `RangeByWeeks` shifts by a week for Sunday,
and the ahead/behind logic has its own history (`3776245`).

Not fixed here: this issue was split out of the calendar work so the two changes
stay reviewable apart.
-->
