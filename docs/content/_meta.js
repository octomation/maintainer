export default {
  '*': { theme: { timestamp: false } },
  index: {
    title: 'Home', type: 'page', display: 'hidden',
    theme: { layout: 'full', sidebar: false, toc: false, breadcrumb: false, pagination: false, copyPage: false },
  },
  documentation: { title: 'Documentation', type: 'page', href: '/quickstart/' },
  releases: { title: 'Changelog', type: 'page', href: '/changelog/' },
  quickstart: 'Get started',
  contributions: 'GitHub contributions',
  vanity: 'Go vanity URLs',
  makefiles: 'Makefiles',
  fetch: 'Fetch repositories',
  status: 'Repository status',
  workspace: 'Workspace',
  releaseHistory: { title: 'Changelog', type: 'separator' },
  changelog: { title: 'Changelog', display: 'children', theme: { pagination: false } },
}
