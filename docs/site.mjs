// The public origin of the site, derived at build time: docs.yml passes the
// GitHub Pages URL as SITE_URL, so a domain change needs no edits here.
// A static export without it would ship localhost links in canonical and OG tags.
if (process.env.TARGET === 'static' && !process.env.SITE_URL) {
  throw new Error('SITE_URL is required for a static export, e.g. SITE_URL=https://maintainer.octolab.org/')
}

export const siteUrl = (process.env.SITE_URL || 'http://localhost:3000').replace(/\/*$/, '/')
