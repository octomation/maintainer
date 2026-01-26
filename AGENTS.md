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

## Releasing

A release is a curated note plus a signed tag. A prerelease (`vX.Y.Z-rcN`) is
a release too: it has its own note and tag and is frozen once published. Once
[MAIN-139](.github/issues/MAIN-139.md) is closed, it stays out of *Latest* and
the tap; until then, don't tag one: it would become *Latest* and update the
tap. Pipeline mechanics, secrets and hooks
are in [.github/workflows/README.md](.github/workflows/README.md); docs
conventions are in [docs/README.md](docs/README.md#keep-the-release-accurate).

### Ground rules

- The maintainer commits, tags and pushes. Stage one self-contained change at a
  time and propose its message in a code block: one line,
  `type(scope): summary`, with issues as `, issue #N` or `, closes #N`; no
  body.
- Product changes first, release commit last: `chore(release): prepare vX.Y.Z`.
  A guide that explains new behavior changes with the product commit.
- The CLI is the source of truth for what the docs and the notes claim.
- A published release is frozen: its note, annotation and image stay as
  tagged. A missed fix ships in the next patch, or in the next candidate
  before the final tag.

### Before you start

A prerelease, the first `v0.2.0-rc1` included, waits until
[MAIN-139](.github/issues/MAIN-139.md) is closed: until GoReleaser publishes a
prerelease tag as a prerelease and skips the tap, a candidate would become
*Latest* and update the Formula and the Cask.

`git fetch` and reconcile `main` with `origin/main`, then:

```sh
for w in ci cd.docs tools doctor; do gh run list -w "$w.yml" -b main -L 1; done
make doctor
(cd docs && npm ci)
```

- The latest `ci` and `cd.docs` runs on `main` are green.
- `tools` is green, or fails only on govulncheck advisories in `tools/` that
  have no fixed version (`Fixed in: N/A`): `cd` installs the tools but doesn't
  scan them.
- `doctor` is green: it mints the tap token from the OctoLab Releaser App,
  which proves the App can publish to the tap. If the App, its installation,
  the `OCTOLAB_RELEASER_*` secrets or `doctor.yml` changed since its last run,
  run it again: `make workflow-doctor REASON="before vX.Y.Z"` dispatches it on
  `main` and waits for the result; `REASON` is optional.
- `make doctor` has no `fail`.

On the candidate commit, with `GITHUB_TOKEN` set:

```sh
make verify
make go-integration-test
make go-dist-check
make release-config-check
```

`make release-config-check` accepts exactly the GoReleaser `brews`
deprecation and nothing else, while the Formula is still published
([MAIN-125](.github/issues/MAIN-125.md)).

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

**Upgrade:** with Homebrew, run `brew update` followed by `brew upgrade --cask octolab/tap/maintainer`; if you still have the Formula, [switch to the Cask](/quickstart/#install). With the install script, pass the tag: `sh -s -- -b ~/.local/bin vX.Y.Z`. With Go, run `go install go.octolab.org/toolset/maintainer@vX.Y.Z`. For a manual installation, use a [vX.Y.Z release archive](https://github.com/octomation/maintainer/releases/tag/vX.Y.Z). Run `maintainer version` to check the installed version.
```

- English; a short slogan ending with a period. v0.1.0, "Find your rhythm",
  predates the rule and stays as published.
- Only what users get: behavior and the compatibility of configuration, state
  and pins, no docs, tools or CI/CD. Explain jargon in user terms; no
  comparisons to release candidates.
- Links are site-relative with a trailing slash. Markdown and HTML, no MDX.
- Once the Formula is retired ([MAIN-125](.github/issues/MAIN-125.md)), drop
  the Formula clause from **Upgrade:**.

A prerelease note, `vX.Y.Z-rcN.md`, is short and self-contained: what to try,
what to report, and links only to pages that are already on the site, not to
the final note. Add it to `changelog/_meta.js` with `display: 'hidden'` and
give it no card.

Then, in the same release commit:

- `changelog/_meta.js`: the new entry on top, `'vX.Y.Z': 'vX.Y.Z — Slogan.'`.
- `changelog/index.mdx`: a new card `vX.Y.Z · LATEST`; the previous one
  becomes `· STABLE`.
- `docs/content/index.mdx`: description, release pill, release band, scope note.
- The current version in `README.md` (the pinned install script example),
  `docs/README.md`, `docs/content/quickstart.md` and the guides. In the content
  table of `docs/README.md`, the new note gets its own row.

```sh
grep -rn 'X\.Y\.W' README.md docs/README.md docs/content docs/public  # only old notes, cards and "since" facts may remain
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

Save the result as RGB PNG (`optimize=True, compress_level=9`).

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

For a prerelease, which needs [MAIN-139](.github/issues/MAIN-139.md) closed,
record the stable channels first:

```sh
gh release view --json tagName -q .tagName
for f in Formula Casks; do gh api repos/octolab/homebrew-tap/contents/$f/maintainer.rb -q .sha; done
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
email: `cd` checks that GitHub verifies the tag before anything is published.

If the push lock is set (`remote.origin.pushurl` is `no_push`), lift it for
this push only and set it again afterwards.

Never push the tag separately. A tag pushed by mistake can be deleted only
before the GitHub release exists; after that, ship the next patch.

### 5. After the push

```sh
gh run watch --exit-status "$(gh run list -w cd.yml -c "$(git rev-parse 'vX.Y.Z^{commit}')" -L 1 --json databaseId -q '.[0].databaseId')"
gh release view vX.Y.Z --json name,isPrerelease,assets -q '.name, .isPrerelease, .assets[].name'
gh api 'repos/octolab/homebrew-tap/commits?per_page=2' -q '.[] | .commit.verification.verified, .commit.message'
```

- `cd` is green. If `gh run watch` says the run is not found, `cd` hasn't
  started yet: run the line again.
- The release has the note's title and body, four archives and
  `checksums.txt`.
- A stable release: the Formula and Cask commits for vX.Y.Z in the tap are
  verified, signed as the App. After the Formula retirement, only the Cask.
- A prerelease, after [MAIN-139](.github/issues/MAIN-139.md): `isPrerelease`
  is `true`, `gh release view --json tagName -q .tagName` still names the
  previous stable release, and both tap files have the recorded `sha`.
- Every channel delivers vX.Y.Z, checked by `maintainer version`,
  `maintainer fetch --help` and `maintainer status --help`, not by a
  successful download alone: the archives, `bin/install` with the tag,
  `go install go.octolab.org/toolset/maintainer@vX.Y.Z` (and `@latest` for a
  stable release) outside a checkout, and Homebrew, with the Formula and the
  Cask in separate installations while both are published:
  `brew update && brew upgrade --cask octolab/tap/maintainer && maintainer version`.
- The site shows the new changelog page and release band once `cd.docs`
  finishes for the release commit.

### When `cd` fails

- **Before the GitHub release exists** (the tag checks, tests or the build;
  `gh release view vX.Y.Z` finds no release): nothing is published. Delete
  the tag with `git push --delete origin vX.Y.Z && git tag -d vX.Y.Z`, fix the
  cause on `main` or in your signing setup, and go back to step 4.
- **A prerelease reached a stable channel** (*Latest* or the tap changed):
  restore the previous stable release as *Latest* and the tap files, then
  file an issue before the next candidate.
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
