import nextra from 'nextra'

const withNextra = nextra({})
const staticExport = process.env.TARGET === 'static'

export default withNextra({
  agentRules: false,
  basePath: process.env.BASE_PATH || '',
  trailingSlash: true,
  // Keep explicit slashes on version-like routes such as /changelog/v0.1.0/.
  skipTrailingSlashRedirect: true,
  ...(staticExport && {
    output: 'export',
    distDir: 'dist',
    images: { unoptimized: true },
  }),
})
