# maintainer documentation

An English, use-case-led site for the v0.1.0 feature set, built on the existing Nextra 4 / Next.js stack.

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
| `content/fetch.md` | Unreleased `fetch`: reconcile local GitHub checkouts |
| `content/status.md` | Unreleased `status`: inspect local repository work |
| `content/workspace.md` | Workspace layout shared by `fetch` and `status` |
| `content/changelog/index.mdx` | Release overview |
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

## Change the domain

The site is served from the custom domain `maintainer.octolab.org`; `octomation.github.io/maintainer/` redirects to it, keeping the path. The build bakes the domain into asset paths and metadata, and changing it in Settings → Pages triggers no rebuild. To change it:

1. Update `pages.cname` in [`.github/settings.json`](../.github/settings.json) (remove it for the default domain) and push.
2. Change Settings → Pages → Custom domain to match.
3. Rebuild: `gh workflow run cd.docs.yml -f reason="domain change"`.

The docs build stops while the two disagree, the deployment is smoke-tested, and the daily [doctor](../.github/workflows/doctor.yml) reports a stale site.

## Keep the release accurate

The CLI implementation is the source of truth: check examples against command help and tests. Closed issues provide context, not a substitute for checking the current code.

The documentation is prepared for v0.1.0. Publish the matching release archives before directing users to its download links. Lead installation instructions with Homebrew and keep the install script, `go install`, and release archives next to it. Archive names and supported platforms must match `.goreleaser.yml`. Do not describe planned features as shipped: a command that is not in a release says so where it is documented.

Release notes describe changes to maintainer's behavior and to the compatibility of its configuration and state. Keep documentation, development tools, CI/CD, and repository inventory changes out of the product changelog. A release note describes its release as published, with examples that reproduce on the built binary. Preserve historical release notes when updating the current installation instructions. How to write a note and cut a release: [AGENTS.md](../AGENTS.md#releasing).

Keep the owner's real collection out of examples: no tokens, private repository names, or local paths; collection numbers appear only as a dated example.

`package.json` pins Zod to `4.1.12` under `nextra-theme-docs` to avoid [a validation regression](https://github.com/shuding/nextra/issues/5036). In the current dependency tree, this also covers the theme's Nextra peer dependency. Keep the exact Zod version: a caret range allows incompatible releases. When upgrading Nextra and its theme, check whether the site builds without this override. See [Nextra’s docs theme guide](https://nextra.site/docs/docs-theme/start) for the underlying structure.

For `speech-rule-engine@4.1.4`, an override replaces its pinned xmldom dependency with `@xmldom/xmldom@^0.9.12` to include the security fixes while allowing later `0.9.x` patches. The lockfile records the resolved version. The current tree resolves `speech-rule-engine@4.0.7`, which does not use that dependency, so the rule takes effect only once an update brings in 4.1.4. It stops applying when speech-rule-engine changes version again; run `npm audit` after that update and remove or revise the rule as needed.

Search is disabled: static search indexing is not configured.
