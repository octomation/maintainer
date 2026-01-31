# maintainer documentation

An English, use-case-led site for the v0.2.0 feature set, built on the existing Nextra 4 / Next.js stack.

## Run locally

Use Node 24 (the CI version). From the repository root:

```sh
./Taskfile docs npm ci
./Taskfile docs dev
```

Open the local URL printed by Next.js, normally <http://localhost:3000>.

Or, from `docs/`, run `npm ci` and `npm run dev`.

## Build

For a production server:

```sh
./Taskfile docs build
./Taskfile docs start
```

For GitHub Pages:

```sh
TARGET=static SITE_URL=https://maintainer.octolab.org/ ./Taskfile docs build
```

Static output is written to `docs/dist/`. `SITE_URL` is required; add `BASE_PATH=/maintainer` only when hosting under a path, as on the default `octomation.github.io/maintainer/`. The [Pages workflow](../.github/workflows/cd.docs.yml) supplies both: from the Pages configuration on main, from `.github/settings.json` for PR previews. Local development remains available while the static build runs.

## Edit the content

| Location | Purpose |
| --- | --- |
| `content/index.mdx` | Product introduction and paths into the guides |
| `content/quickstart.md` | Installation and a first calendar workflow |
| `content/contributions.md` | GitHub contribution calendar commands |
| `content/vanity.md` | Go vanity import page generation |
| `content/makefiles.md` | Makefile bundling |
| `content/fetch.md` | `fetch`: reconcile local GitHub checkouts |
| `content/status.md` | `status`: inspect local repository work |
| `content/workspace.md` | Workspace layout shared by `fetch` and `status` |
| `content/agents.md` | The agent skill: install, keep in step, catalog |
| `content/changelog/index.mdx` | Release overview |
| `content/changelog/v0.2.0.md` | Fetch and status release notes |
| `content/changelog/v0.1.0.md` | First stable release notes |
| `content/**/_meta.js` | Navigation order and labels |
| `navigation.mjs` | Release groups in the sidebar; keeps published release URLs unchanged |
| `app/globals.css` | Shared colors, responsive layouts, and landing page styles; kept identical to indexit |
| `app/brand.css` | maintainer colors over the shared styles |
| `app/layout.jsx` | Site identity and Nextra theme |
| `app/[[...mdxPath]]/page.jsx` | Page-specific titles and social metadata |
| `public/` | Favicon and social preview |

The public origin in metadata comes from `SITE_URL` (see `site.mjs`); the Pages workflow sets it from the Pages configuration. Internal links use Next.js/Nextra and keep the configured base path.

Get started lives at `/quickstart/`; the guides stay at the top level, such as `/contributions/` and `/fetch/`. Published release notes link to these URLs, so keep them. The sidebar groups the guides through `_meta.js`.

Changelog is a section of the same sidebar, with All releases first. `navigation.mjs` groups release pages into expandable `v0.1.x`, `v0.2.x`, and later branches, newest first. Add each release note and its label to `content/changelog/_meta.js` as before; the version group is created automatically. A prerelease note, such as `v0.2.0-rc1`, is not grouped: add it with `display: 'hidden'`, so its page exists for the GitHub release but the sidebar shows stable releases only. Keep the notes at `content/changelog/{tag}.md` so published links and the release workflow stay valid. Top-level navbar links use separate `type: 'page'` aliases; the changelog folder uses `display: 'children'` to keep the full sidebar available across guides and releases.

## Social preview image

`public/og.png`, 1733×908, is the social preview of the home page; other pages share a link without an image (`app/[[...mdxPath]]/page.jsx`). On warm, lightly textured off-white paper, it shows the `maintainer` wordmark in heavy lowercase ink with a green dot over the first i, the slogan "Keep the little things moving.", the monospaced line "See your pace. Plan your next contribution." and the green version label `v0.2.0` at the bottom left. On the right, a thin terminal frame holds a seven-row contribution calendar with one suggested square.

Each release patches only the changed characters of the version label, so the rest of the image stays byte-for-byte as it was: see [AGENTS.md](../AGENTS.md#2-social-preview-image), which also has the diff check. Coordinates are pixels from the top left as (x, y, width, height), right and bottom edges exclusive:

| Part | Geometry |
| --- | --- |
| Visible label | (100, 739, 135, 29) |
| Editing box | (97, 735, 141, 37) |
| Cell boundaries | x=97, 121, 144, 168, 191, 215, 238: about 23.5 px, alternating 24 and 23 px |
| Baseline | y=768 |
| Clean background donor strip | (260, 735, 141, 37) |
| Label type | monospaced, close to Overpass Mono at about 40 px, weight 450, `#4d7c0f` |

Regenerate the image in full only when the wordmark, the slogan, the monospaced line or the brand colors change. Give Codex `gpt-6-astra` at `high` effort with image generation the current image, or the indexit one it was modeled on (`octopot/indexit`, `docs/public/og.png`), as the reference, with this prompt, updated to the new text:

```text
Use case: ads-marketing.
Create the final social preview banner for maintainer, a sibling CLI to indexit. The provided indexit image is a visual reference: closely mirror its layout, typographic character, warm off-white lightly textured paper, dark ink, thin line-art and generous whitespace, but replace its content as specified below. Wide canvas, aspect ratio 1733:908. Entire image is the artwork, no mockup or border.

Use positions in a 1733 x 908 coordinate system. Keep the left margin around x=99. Left typography occupies roughly 65% of the width, right illustration roughly 30%. Wordmark around y=220..462: "maintainer", spelled m a i n t a i n e r, all lowercase, very heavy geometric sans like the reference. Fit the longer word within x=99..1170 without squeezing the letters unnaturally. Ink #17302e, except the dot over its first i is green #4d7c0f. Only one accent dot in the wordmark.
Under the wordmark, a bold sans slogan on ONE line, exact text: "Keep the little things moving." Use dark ink #17302e, fit in the left area, around y=525..596.
Below that, ONE line in a monospaced font, dark ink #17302e, exact text: "See your pace. Plan your next contribution." Around x=99, y=644..675. Keep comfortably separated from the illustration.
Bottom left version label, exact text: "v0.2.0". Match reference label position (visible glyphs around x=100, y=739..768) and size (approximately 40px font, approximately 23px fixed character cells). Use crisp regular/medium monospaced type close to Overpass Mono weight 450, green #4d7c0f. All six characters on one baseline, evenly spaced. Keep it isolated on a clean patch of the same off-white paper with no nearby marks so future releases can replace just the glyphs.

Right-hand illustration around x=1330..1680: thin dark ink terminal frame, a green > prompt, and a small contribution-calendar grid INSIDE the terminal frame. Grid must have exactly SEVEN rows of evenly spaced square cells, about eight columns. Most cells empty or faint neutral outlines, some filled #4d7c0f, a few lime #d5f45a highlights. One suggested-next-contribution square has a clear extra outline with a short dotted leader. Include the reference's sparse green accent squares, dotted vertical leaders, and minimal thin framing strokes above and below. No letters, numerals or data labels in illustration; only > prompt, squares, lines and dots. Keep its visual weight lighter than the wordmark. Optional thin green arc at lower right like the reference.

Render all four text elements exactly as quoted, sharp and legible. No other text, no indexit wording, no logos of GitHub or other brands, no people, no watermark. Flat artwork, RGB appearance, opaque off-white background.
```

The first pass of the current image drew six calendar rows; a second edit pass asked for exactly one change, a seventh row below the sixth, and to preserve everything else. Check every text element and the row count before accepting a result.

Post-process with Pillow: convert to RGB, resize to exactly 1733×908 (LANCZOS; the tool returned 1732×908), and save as PNG with `optimize=True, compress_level=9`, no alpha or palette. Keep the file under 1 MB: if it is larger, round each channel to an even value, at most one level of 255, and save again. Then measure the label anew, update the geometry above and in AGENTS.md, and check `width`, `height` and `alt` in `app/[[...mdxPath]]/page.jsx`.

## Change the domain

The site is served from the custom domain `maintainer.octolab.org`; `octomation.github.io/maintainer/` redirects to it, keeping the path. The build bakes the domain into asset paths and metadata, and changing it in Settings → Pages triggers no rebuild. To change it:

1. Update `pages.cname` in [`.github/settings.json`](../.github/settings.json) (remove it for the default domain) and push.
2. Change Settings → Pages → Custom domain to match.
3. Rebuild: `gh workflow run cd.docs.yml -f reason="domain change"`.

The docs build stops while the two disagree, the deployment is smoke-tested, and the daily [doctor](../.github/workflows/doctor.yml) reports a stale site.

## Keep the release accurate

The CLI implementation is the source of truth: check examples against command help and tests. Closed issues provide context, not a substitute for checking the current code.

The documentation is prepared for v0.2.0. Publish the matching release archives before directing users to its download links. Lead installation instructions with Homebrew and keep the install script, `go install`, and release archives next to it. Archive names and supported platforms must match `.goreleaser.yml`. Do not describe planned features as shipped: a command that is not in a release says so where it is documented.

Release notes describe changes to maintainer's behavior and to the compatibility of its configuration and state. Keep documentation, development tools, CI/CD, and repository inventory changes out of the product changelog. A release note describes its release as published, with examples that reproduce on the built binary. Preserve historical release notes when updating the current installation instructions. How to write a note and cut a release: [AGENTS.md](../AGENTS.md#releasing).

Keep the owner's real collection out of examples: no tokens, private repository names, or local paths; collection numbers appear only as a dated example.

`package.json` pins Zod to `4.1.12` under `nextra-theme-docs` to avoid [a validation regression](https://github.com/shuding/nextra/issues/5036). In the current dependency tree, this also covers the theme's Nextra peer dependency. Keep the exact Zod version: a caret range allows incompatible releases. When upgrading Nextra and its theme, check whether the site builds without this override. See [Nextra’s docs theme guide](https://nextra.site/docs/docs-theme/start) for the underlying structure.

For `speech-rule-engine@4.1.4`, an override replaces its pinned xmldom dependency with `@xmldom/xmldom@^0.9.12` to include the security fixes while allowing later `0.9.x` patches. The lockfile records the resolved version. The current tree resolves `speech-rule-engine@4.0.7`, which does not use that dependency, so the rule takes effect only once an update brings in 4.1.4. It stops applying when speech-rule-engine changes version again; run `npm audit` after that update and remove or revise the rule as needed.

Search is disabled: static search indexing is not configured.
