# Agent guide

## Features

A larger feature goes from its issue to the code to the user guide.

- **The issue comes first.** Before you implement it, its issue in
  `.github/issues/MAIN-N.md` states the expected behavior, the limits and the
  acceptance criteria. When the requirements change, update the issue before
  the code. The issue sets the requirements: don't rewrite it to fit a defect
  or a limit of the implementation.
- **The guide finishes it.** A feature is done when `docs/content/` describes
  it for users: what it does, how to use it, and an example of the result.
  Update the command's page or add one; mention only the flags and limits
  needed to use it, and check the examples against the code and `--help`. The
  guide changes in the product commit ([ground rules](#ground-rules)); what a
  guide may claim is in
  [docs/README.md](docs/README.md#keep-the-release-accurate).
- A guide stands on its own: no development history, implementation plans,
  test reports, internals or promises of unreleased behavior.

## Hub frontend

The pages of the local service live in `web/` (Preact, TypeScript, Vite; the
built `web/dist` is committed and embedded in the binary). Their look is a
contract: [web/DESIGN.md](web/DESIGN.md) holds the tokens and the component
rules, in the design.md format, and `npx @google/design.md lint web/DESIGN.md`
must report no errors. Change a token there before the code, keep both themes
and the 390 px layout, and run `make web-check` before handing over;
[web/README.md](web/README.md) explains the layout of the sources.

## Local services

A local service listens on loopback only, on `127.0.0.1` and `::1`, and
answers at `http://<name>.localhost:<port>/`; `brew services` runs it as the
user, without root. Each one has its own port, recorded here and in the
AGENTS.md of the other repository:

| Port | Service |
| --- | --- |
| 17840 | indexit |
| 17841 | maintainer |

A new service takes the next free port and adds it to both tables.

## Releasing

A release is a curated note plus a signed tag. A stable release also publishes
the agent skill to the [octolab/skills](https://github.com/octolab/skills)
catalog. A prerelease (`vX.Y.Z-rcN`) is a release too: it has its own note and
tag and is frozen once published. GoReleaser publishes it as a prerelease that
never becomes *Latest* and leaves the tap as it is, and the catalog keeps
stable releases only. Pipeline mechanics, secrets and hooks
are in [.github/workflows/README.md](.github/workflows/README.md); docs
conventions are in [docs/README.md](docs/README.md#keep-the-release-accurate).

### Ground rules

- The maintainer commits, tags and pushes. Stage one self-contained change at a
  time and propose its message in a code block: one line,
  `type(scope): summary`, with issues as `, issue #N` or `, closes #N`; no
  body.
- Product changes first, release commit last: `chore(release): prepare vX.Y.Z`.
  A guide that explains new behavior changes with the product commit, and so
  does the agent skill in `skills/maintainer/`; `go test ./skills` keeps its
  examples and command reference in step with the CLI, and
  `go generate ./skills` rewrites the reference.
- The CLI is the source of truth for what the docs, the notes and the skill
  claim.
- A published release is frozen: its note, annotation, image and catalog entry
  stay as tagged. A missed fix ships in the next patch, or in the next candidate
  before the final tag.

### Before you start

`git fetch` and reconcile `main` with `origin/main`, then:

```sh
for w in ci cd.docs tools doctor; do gh run list -w "$w.yml" -b main -L 1; done
gh run list -R octolab/skills -w ci.yml -b main -L 1
make doctor
(cd docs && npm ci)
make web-check
```

- The latest `ci` and `cd.docs` runs on `main` are green.
- `tools` is green, or fails only on govulncheck advisories in `tools/` that
  have no fixed version (`Fixed in: N/A`): `cd` installs the tools but doesn't
  scan them.
- `doctor` is green: it mints the tap and catalog tokens from the OctoLab
  Releaser App, which proves the App can publish to both. If the App, its
  installations, the `OCTOLAB_RELEASER_*` secrets or `doctor.yml` changed
  since its last run, run it again:
  `make workflow-doctor REASON="before vX.Y.Z"` dispatches it on `main` and
  waits for the result; `REASON` is optional.
- `ci` in `octolab/skills` is green, and its `catalog.json` registers
  `maintainer`: the publish action refuses an invalid catalog and an
  unregistered tool.
- `make doctor` has no `fail`.

On the candidate commit, with `GITHUB_TOKEN` set:

```sh
make verify
make go-integration-test
make go-dist-check
make release-config-check
```

`make release-config-check` accepts exactly the GoReleaser `brews`
deprecation and nothing else: `brews` ships the CLI and the macOS service as
one Formula, and the Cask is no longer published
([MAIN-125](.github/issues/MAIN-125.md)). Existing Cask users need the
one-time move in the
[service guide](docs/content/service.md#move-from-the-cask). Published notes
keep their original Cask commands.

### 1. Changelog

Write `docs/content/changelog/vX.Y.Z.md`; it becomes the GitHub release:

```md
---
title: vX.Y.Z — Slogan.
description: "One sentence on what changed for the user."
---

# vX.Y.Z — Slogan.

Lead paragraph: what the release does, in terms of commands and results.

## A section per user-facing change

- Short, concrete bullets.

[Inspect your checkouts →](/status/)

## Known limitations

- What the release does not do yet, in user terms.

**Upgrade:** with Homebrew, run `brew update` followed by `brew upgrade --formula octolab/tap/maintainer`. With the install script, pass the tag: `sh -s -- -b ~/.local/bin vX.Y.Z`. With Go, run `go install go.octolab.org/toolset/maintainer@vX.Y.Z`. For a manual installation, use a [vX.Y.Z release archive](https://github.com/octomation/maintainer/releases/tag/vX.Y.Z). Run `maintainer version` to check the installed version.
```

- English; a short slogan ending with a period. v0.1.0, "Find your rhythm",
  predates the rule and stays as published.
- Only what users get: behavior and the compatibility of configuration, state
  and pins, no docs, tools or CI/CD. Explain jargon in user terms; no
  comparisons to release candidates.
- Links are site-relative with a trailing slash. Markdown and HTML, no MDX.
- After v0.2.1, end the **Upgrade:** paragraph with
  `If you use the agent skill, [replace it](/agents/#keep-it-in-step) with the new binary’s copy.`

A prerelease note, `vX.Y.Z-rcN.md`, is short and self-contained: what to try,
what to report, and links only to pages that are already on the site, not to
the final note. Add it to `changelog/_meta.js` with `display: 'hidden'` and
give it no card. Of the list below, a prerelease changes only
`skills/maintainer/SKILL.md`: the current version elsewhere stays the stable
one.

Then, in the same release commit:

- `changelog/_meta.js`: the new entry on top, `'vX.Y.Z': 'vX.Y.Z — Slogan.'`.
- `changelog/index.mdx`: a new card `vX.Y.Z · LATEST`; the previous one
  becomes `· STABLE`.
- `docs/content/index.mdx`: description, release pill, release band, scope note.
- The current version in `README.md` (the pinned install script example),
  `docs/README.md`, `docs/content/quickstart.md` and the guides, catalog tags
  such as `maintainer--vX.Y.Z` included. In the content table of
  `docs/README.md`, the new note gets its own row.
- `skills/maintainer/SKILL.md`: `metadata.version` is the tag without `v`;
  `tool-version-range`, the `compatibility` line and the ranges in the skill's
  text become `>=0.Y.Z <0.(Y+1).0` before 1.0, `>=X.Y.Z <(X+1).0.0` after,
  and the exact tag without `v` for a prerelease, such as `=0.3.0-rc1`.
  `make release-check` rejects another version, `go test ./skills` mismatched
  ranges.

```sh
grep -rn 'X\.Y\.W' README.md docs/README.md docs/content docs/public skills  # only old notes, cards and "since" facts may remain
go test ./skills
node .github/scripts/release.mjs render vX.Y.Z --site-url https://maintainer.octolab.org/ --out "$(mktemp -d)/notes.md"
(cd docs && TARGET=static SITE_URL=https://maintainer.octolab.org/ npx next build)
```

Check the sidebar: All releases first, then version groups with the new note;
no prerelease note in it.

### 2. Social preview image

`docs/public/og.png` shows the version at the bottom left. Patch only the
changed characters of the label, don't regenerate the image or resize or
requantize it: work on an untouched RGB copy at native resolution.
Coordinates are pixels from the top left as (x, y, width, height); the label
geometry is in [docs/README.md](docs/README.md#social-preview-image).

- The label is monospaced, close to Overpass Mono at about 40 px, weight 450,
  `#4d7c0f`, on the baseline y=768. Its editing box is (97, 735, 141, 37), six
  cells of about 23.5 px: the boundaries are x=97, 121, 144, 168, 191, 215,
  238, so the cells alternate 24 and 23 px; don't step by a fixed 24 px.
- Cover each changed cell with background copied from the clean donor strip
  (260, 735, 141, 37) at the same rows: for the cell (x, 735, w, 37), copy
  (260 + x - 97, 735, w, 37). The background is textured; keep unchanged cells.
- Copy a new glyph from the untouched label when it is there, with its
  antialiased fringe, shifted by whole cells (two cells are 47 px) at the same
  y, through a foreground mask that keeps the destination paper. Otherwise
  draw it at the baseline in the font above, rendered at 4× and downsampled
  as a patch, after matching an existing `0` or `1` on a scratch crop: the
  lettering was generated, the font name is an approximation.
- A longer version extends the editing box by whole cells into the blank
  space to the right; update the coordinates here and in docs/README.md.

Compare the label at 4× and diff the images: only the editing box may change.
With Pillow, `before.png` being the untouched original:

```python
from PIL import Image, ImageChops
a = Image.open("before.png").convert("RGB")
b = Image.open("docs/public/og.png").convert("RGB")
assert a.size == b.size == (1733, 908)
diff = ImageChops.difference(a, b)
bounds = diff.getbbox()
assert bounds is not None, "No version change"
x0, y0, x1, y1 = bounds
assert 97 <= x0 < x1 <= 238 and 735 <= y0 < y1 <= 772, bounds
outside = diff.copy()
outside.paste((0, 0, 0), (97, 735, 238, 772))
assert outside.getbbox() is None
for name, im in [("before", a), ("after", b)]:
    im.crop((85, 725, 255, 782)).resize((680, 228), Image.Resampling.NEAREST).save(name + "-label-4x.png")
print("Only the version label changed:", bounds)
```

Save the result as RGB PNG (`optimize=True, compress_level=9`). The binary
embeds this file (`main.go`) and `maintainer service` answers with it, so the
release build carries the new label; there is no second copy to update.

A prerelease doesn't change the image. Regenerate it in full only when the
wordmark, the slogan, the monospaced line or the brand colors change: with
Codex `gpt-6-astra` at `high` effort and image generation, from the prompt in
[docs/README.md](docs/README.md#social-preview-image). Then post-process it
to 1733×908 RGB as described there, measure the label again, and update the
coordinates in both files and `width`, `height` and `alt` in
`docs/app/[[...mdxPath]]/page.jsx` if they changed.

### 3. Tag annotation

Follow the previous tag, `git tag -l --format='%(contents)' vX.Y.W`:

```text
vX.Y.Z — Slogan.

One or two sentences on what the release does, wrapped
at 72 columns.

- lowercase imperative bullets without trailing periods

Release notes: docs/content/changelog/vX.Y.Z.md
```

Same scope as the note. Write it to a file outside the repository:

```sh
annotation="$(mktemp -d)/annotation.txt"
```

### 4. Tag, check, push

For a prerelease, record the stable channels first:

```sh
gh release view --json tagName -q .tagName
gh api repos/octolab/homebrew-tap/contents/Formula/maintainer.rb -q .sha
```

Then:

```sh
git tag -s vX.Y.Z -F "$annotation"
git tag -v vX.Y.Z
make release-check TAG=vX.Y.Z
git push --atomic origin main vX.Y.Z
```

Run them on `main` with the release commit checked out: `make release-check`
checks the tag against the branch you are about to push, as the `pre-push`
hook does, together with tidy modules and the GoReleaser config.

`git tag -v` must report a good signature from a key GitHub knows for your
email: `cd` checks that GitHub verifies the tag before anything is published,
and the catalog publishes only from a verified tag.

If the push lock is set (`remote.origin.pushurl` is `no_push`), lift it for
this push only and set it again afterwards.

Never push the tag separately. A tag pushed by mistake can be deleted only
before the GitHub release exists; after that, ship the next patch.

### 5. After the push

```sh
gh run watch --exit-status "$(gh run list -w cd.yml -c "$(git rev-parse 'vX.Y.Z^{commit}')" -L 1 --json databaseId -q '.[0].databaseId')"
gh release view vX.Y.Z --json name,isPrerelease,assets -q '.name, .isPrerelease, .assets[].name'
gh api 'repos/octolab/homebrew-tap/commits?per_page=1' -q '.[] | .commit.verification.verified, .commit.message'
gh api repos/octolab/skills/commits/maintainer--vX.Y.Z -q '.commit.verification.verified, .commit.message'
gh run list -R octolab/skills -w ci.install.yml -b maintainer--vX.Y.Z -L 1
```

- `cd` is green; for a prerelease, its skill job is skipped. If `gh run watch`
  says the run is not found, `cd` hasn't started yet: run the line again.
- The release has the note's title and body, four archives and
  `checksums.txt`.
- A stable release: the Formula commit for vX.Y.Z in the tap is verified,
  signed as the App. After the first release with the service, delete
  `Casks/maintainer.rb` from `octolab/homebrew-tap` as a separate reviewed
  change ([MAIN-125](.github/issues/MAIN-125.md)): GoReleaser does not delete
  the old Cask.
- A prerelease: `isPrerelease` is `true`,
  `gh release view --json tagName -q .tagName` still names the previous stable
  release, and the Formula has the recorded `sha`.
- Every channel delivers vX.Y.Z, checked by `maintainer version`,
  `maintainer fetch --help`, `maintainer status --help` and
  `maintainer service run --help`, not by a successful download alone: the
  archives, `bin/install` with the tag,
  `go install go.octolab.org/toolset/maintainer@vX.Y.Z` (and `@latest` for a
  stable release) outside a checkout, and Homebrew:
  `brew update && brew upgrade --formula octolab/tap/maintainer && maintainer version`.
- On macOS, after `brew services restart octolab/tap/maintainer`,
  `curl -fsS http://maintainer.localhost:17841/status -o /dev/null -w '%{http_code}'`
  prints `200` on the release commit.
- A stable release only: the catalog tag `maintainer--vX.Y.Z` points at a
  verified commit, and `ci (install)` there is green for that tag.
- A stable release only: the catalog plugins are updated, then
  `maintainer skill status --agent claude-code,codex --exact` passes:

  ```sh
  claude plugin marketplace update octolab && claude plugin update maintainer@octolab
  codex plugin marketplace upgrade octolab && codex plugin add maintainer@octolab
  ```

  The skill comes from the catalog, not from `maintainer skill install`; `ci`
  checks that install on every change. A copy that `status` reports next to
  the plugin is a leftover: remove it, the agent would load both.
- The site shows the new changelog page and release band once `cd.docs`
  finishes for the release commit.

### When `cd` fails

- **Before the GitHub release exists** (the tag checks, tests or the build;
  `gh release view vX.Y.Z` finds no release): nothing is published. Delete
  the tag with `git push --delete origin vX.Y.Z && git tag -d vX.Y.Z`, fix the
  cause on `main` or in your signing setup, and go back to step 4.
- **A prerelease reached a stable channel** (*Latest* or the tap changed):
  restore the previous stable release as *Latest* and the Formula, then
  file an issue before the next candidate.
- **Only the skill job failed**: the release and the tap are out and stay.
  Fix the cause, for example the App's installation on `octolab/skills` or
  the tool's entry in its `catalog.json`, and run the failed job again:
  `gh run rerun <run-id> --failed`.
- **Anything else after the release exists**: it is frozen; ship the fix in
  the next patch.

### 6. Announcement

Propose a row for "Latest contributions" in the
[kamilsk/kamilsk](https://github.com/kamilsk/kamilsk) profile README:

```md
| 2026-09-25   | 👨‍🔧 [maintainer][]           | [v0.1.0][om010]                   | Contribution calendar planning in CLI     |

[om010]:  https://github.com/octomation/maintainer/releases/tag/v0.1.0
```

Newest first, one row per release day; a prerelease gets no row. The
reference label is `o` + project initials + version digits. Notes is a short
phrase, no longer than the existing ones, saying what the user gets.
