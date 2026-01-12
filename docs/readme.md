# maintainer documentation

The public site for maintainer, built with Nextra 4 on Next.js.

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

Static output is written to `docs/dist/`. `SITE_URL` is required; add `BASE_PATH=/maintainer` only when hosting under a path, as on the default `octomation.github.io/maintainer/`. The [docs workflow](../.github/workflows/docs.yml) supplies both: from Pages on main, from `.github/settings.json` for PR previews.

## Edit the content

| Location | Purpose |
| --- | --- |
| `content/index.mdx` | Overview and paths into the guides |
| `content/quickstart.md` | Installation and a first calendar workflow |
| `content/contributions.md` | GitHub contribution calendar commands |
| `content/vanity.md` | Go vanity import page generation |
| `content/makefiles.md` | Makefile bundling |
| `content/changelog/index.md` | Short release history |
| `content/changelog/vX.Y.Z.md` | Release notes, one per tag |
| `content/**/_meta.js` | Navigation order and labels |
| `app/globals.css` | Shared styles |
| `app/layout.jsx` | Site identity and Nextra theme |
| `app/[[...mdxPath]]/page.jsx` | Page titles and social metadata |
| `public/` | Favicon |

The public origin in metadata comes from `SITE_URL` (see `site.mjs`). Internal links use Next.js/Nextra and keep the configured base path.

## Write a release note

A release is a curated note plus a tag. Add `content/changelog/vX.Y.Z.md` with `title` and `description` frontmatter and an H1 equal to the title; Markdown and HTML, no MDX, site-relative links. The release workflow validates the note of the tag and renders it as the GitHub release body. Write the note when the release is ready, see [releases](../.github/workflows/README.md#cd).

## Change the domain

The site is served from the custom domain `maintainer.octolab.org`; `octomation.github.io/maintainer/` redirects to it, keeping the path. The build bakes the domain into asset paths and metadata, and changing it in Settings → Pages triggers no rebuild. To change it:

1. Update `pages.cname` in [`.github/settings.json`](../.github/settings.json) (remove it for the default domain) and push.
2. Change Settings → Pages → Custom domain to match.
3. Rebuild: `gh workflow run docs.yml -f reason="domain change"`.

The docs build stops while the two disagree, the deployment is smoke-tested, and the daily [doctor](../.github/workflows/doctor.yml) reports a stale site.

## Keep it accurate

The CLI implementation is the source of truth: check examples against command help and tests. The v0.1.0 note describes the release being prepared, not an already published tag; update its status when the tag ships. Archive names and platforms must match `.goreleaser.yml`.

`package.json` pins Zod to `4.1.12` for Nextra and its theme to avoid [a validation regression](https://github.com/shuding/nextra/issues/5036). Revisit those overrides when upgrading Nextra. Search is disabled: static search indexing is not configured.
