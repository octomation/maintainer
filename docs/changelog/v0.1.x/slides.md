# v0.1.x changes

The core of the changes is the new `maintainer github contribution` command.
It supports the [GitHub Contributions Calendar][calendar].

**Motivation:** for me, it provides gamification mechanics -
every time I contribute to the project, I align the contribution calendar.
It's simple and yet powerful.

[calendar]: https://docs.github.com/en/account-and-profile/setting-up-and-managing-your-github-profile/managing-contribution-graphs-on-your-profile/viewing-contributions-on-your-profile#contributions-calendar

---

## Tips and tricks

### Contribution Suggestion

```bash
commit() {
    timestamp=$(maintainer github contribution suggest --short
        "$(git --no-pager log -1 --format="%as")"
    )
    COMMITTER_DATE="${timestamp}" git commit --date="${timestamp}" -m "${*:1}"
)

git config alias.cmm "!commit"
git cmm commit to the past
```
