---
id: 148
database_id: 1773169593
node_id: I_kwDOE2M9Zc5psGu5
status: open
title: "github: contribution: edge case for contrib suggestion"
labels: [scope: code, type: bug, severity: critical, impact: medium, effort: medium]
url: https://github.com/octomation/maintainer/issues/148
created_at: 2023-06-25T09:20:51Z
updated_at: 2023-06-25T09:20:52Z
---

# github: contribution: edge case for contrib suggestion

**Details**

```bash
$ which maintainer
/Users/ksamigullin/go/bin/maintainer

$ git contrib dev: add init task
recovered: assertion is not a true
---
unexpected panic occurred
go.octolab.org/safe.Do.func2
        /Users/ksamigullin/go/pkg/mod/go.octolab.org@v0.12.2/safe/do.go:26
runtime.gopanic
        /opt/homebrew/Cellar/go/1.20.5/libexec/src/runtime/panic.go:884
go.octolab.org/toolset/maintainer/internal/pkg/assert.True
        /Users/ksamigullin/Development/public/octomation/maintainer/internal/pkg/assert/assert.go:33
go.octolab.org/toolset/maintainer/internal/pkg/time.NewRange
        /Users/ksamigullin/Development/public/octomation/maintainer/internal/pkg/time/range.go:11
go.octolab.org/toolset/maintainer/internal/pkg/time.Range.Since
        /Users/ksamigullin/Development/public/octomation/maintainer/internal/pkg/time/range.go:105
go.octolab.org/toolset/maintainer/internal/command/github/contribution.Suggest.func1
        /Users/ksamigullin/Development/public/octomation/maintainer/internal/command/github/contribution/suggest.go:46
github.com/spf13/cobra.(*Command).execute
        /Users/ksamigullin/go/pkg/mod/github.com/spf13/cobra@v1.7.0/command.go:940
github.com/spf13/cobra.(*Command).ExecuteC
        /Users/ksamigullin/go/pkg/mod/github.com/spf13/cobra@v1.7.0/command.go:1068
github.com/spf13/cobra.(*Command).Execute
        /Users/ksamigullin/go/pkg/mod/github.com/spf13/cobra@v1.7.0/command.go:992
github.com/spf13/cobra.(*Command).ExecuteContext
        /Users/ksamigullin/go/pkg/mod/github.com/spf13/cobra@v1.7.0/command.go:985
main.main.func1
        /Users/ksamigullin/Development/public/octomation/maintainer/main.go:48
go.octolab.org/safe.Do
        /Users/ksamigullin/go/pkg/mod/go.octolab.org@v0.12.2/safe/do.go:29
main.main
        /Users/ksamigullin/Development/public/octomation/maintainer/main.go:48
runtime.main
        /opt/homebrew/Cellar/go/1.20.5/libexec/src/runtime/proc.go:250
runtime.goexit
        /opt/homebrew/Cellar/go/1.20.5/libexec/src/runtime/asm_arm64.s:1172
fatal: invalid date format:

$ git --no-pager log -1 
commit dca8015b0eb4f750fdd6e14221afef59accf962d (HEAD -> main)
Author: Kamil Samigullin <kamil@samigullin.info>
Date:   Mon Jun 26 08:35:26 2023 +0300

    chore: up go version

$ datetime
2023-06-25 12:15:27 +0300
```
