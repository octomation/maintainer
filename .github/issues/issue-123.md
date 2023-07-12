---
id: 123
database_id: 1652194598
node_id: I_kwDOE2M9Zc5ien0m
status: closed
title: "github: contribution: suggest and lookup paniced on Sunday"
labels: [scope: code, type: bug, severity: critical, impact: high, effort: medium]
url: https://github.com/octomation/maintainer/issues/123
created_at: 2023-04-03T14:31:28Z
updated_at: 2023-04-05T18:50:14Z
---

# github: contribution: suggest and lookup paniced on Sunday

**Details**

```bash
$ datetime
2023-04-02 09:09:53 +0300

$ maintainer github contribution suggest /-20
recovered: assertion is not a true
---
unexpected panic occurred
go.octolab.org/safe.Do.func2
	go.octolab.org@v0.12.2/safe/do.go:26
runtime.gopanic
	runtime/panic.go:884
go.octolab.org/toolset/maintainer/internal/pkg/assert.True
	go.octolab.org/toolset/maintainer/internal/pkg/assert/assert.go:33
go.octolab.org/toolset/maintainer/internal/pkg/time.Range.ExpandRight
	go.octolab.org/toolset/maintainer/internal/pkg/time/range.go:115
go.octolab.org/toolset/maintainer/internal/command/github/exec.Contribution.func1
	go.octolab.org/toolset/maintainer/internal/command/github/exec/contribution.go:33
github.com/spf13/cobra.(*Command).execute
	github.com/spf13/cobra@v1.6.1/command.go:916
github.com/spf13/cobra.(*Command).ExecuteC
	github.com/spf13/cobra@v1.6.1/command.go:1044
github.com/spf13/cobra.(*Command).Execute
	github.com/spf13/cobra@v1.6.1/command.go:968
github.com/spf13/cobra.(*Command).ExecuteContext
	github.com/spf13/cobra@v1.6.1/command.go:961
main.main.func1
	go.octolab.org/toolset/maintainer/main.go:48
go.octolab.org/safe.Do
	go.octolab.org@v0.12.2/safe/do.go:29
main.main
	go.octolab.org/toolset/maintainer/main.go:48
runtime.main
	runtime/proc.go:250
runtime.goexit
	runtime/asm_arm64.s:1172

$ maintainer github contribution lookup /-20
recovered: assertion is not a true
---
unexpected panic occurred
go.octolab.org/safe.Do.func2
	go.octolab.org@v0.12.2/safe/do.go:26
runtime.gopanic
	runtime/panic.go:884
go.octolab.org/toolset/maintainer/internal/pkg/assert.True
	go.octolab.org/toolset/maintainer/internal/pkg/assert/assert.go:33
go.octolab.org/toolset/maintainer/internal/pkg/time.Range.Shift
	go.octolab.org/toolset/maintainer/internal/pkg/time/range.go:122
go.octolab.org/toolset/maintainer/internal/command/github.Contribution.func2
	go.octolab.org/toolset/maintainer/internal/command/github/contribution.go:185
github.com/spf13/cobra.(*Command).execute
	github.com/spf13/cobra@v1.6.1/command.go:916
github.com/spf13/cobra.(*Command).ExecuteC
	github.com/spf13/cobra@v1.6.1/command.go:1044
github.com/spf13/cobra.(*Command).Execute
	github.com/spf13/cobra@v1.6.1/command.go:968
github.com/spf13/cobra.(*Command).ExecuteContext
	github.com/spf13/cobra@v1.6.1/command.go:961
main.main.func1
	go.octolab.org/toolset/maintainer/main.go:48
go.octolab.org/safe.Do
	go.octolab.org@v0.12.2/safe/do.go:29
main.main
	go.octolab.org/toolset/maintainer/main.go:48
runtime.main
	runtime/proc.go:250
runtime.goexit
	runtime/asm_arm64.s:1172
```
