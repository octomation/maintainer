---
id: 57
database_id: 1272193691
node_id: I_kwDOE2M9Zc5L1CKb
status: closed
title: "github: contribution: refactor logic of suggest command"
labels: [scope: docs, scope: code]
url: https://github.com/octomation/maintainer/issues/57
created_at: 2022-06-15T12:54:26Z
updated_at: 2022-06-15T14:23:05Z
---

# github: contribution: refactor logic of suggest command

**Motivation:** there is a possibility to simplify code

https://github.com/kamilsk/dotfiles/blob/d19edea4e4a08a325acece3ddb41f5c5312a0133/bin/git_commit#L40-L43

**To do**

- [x] use view option instead of a lot of args
- [x] use `--delta` to show in format `-123d`
- [x] solve `TODO:magic replace by params`
- [x] simplify code in `dotfiles`
  - [x] use current git history to make a decision about valid date

**Related to**

- #56 
- kamilsk/dotfiles/issues/320
