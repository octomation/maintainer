import { Footer, Layout, Navbar } from 'nextra-theme-docs'
import { Head } from 'nextra/components'
import { getPageMap } from 'nextra/page-map'
import 'nextra-theme-docs/style.css'
import './globals.css'
import './brand.css'
import { siteUrl } from '../site.mjs'
import { groupReleases } from '../navigation.mjs'

export const metadata = {
  title: { default: 'maintainer — Small tools for open source upkeep', template: '%s · maintainer' },
  description: 'See your open source contribution pace and plan your next session from the terminal.',
  metadataBase: new URL(siteUrl),
  icons: { icon: `${process.env.BASE_PATH || ''}/icon.svg` },
}

function withExportLinks(items) {
  return items.map(item => ({
    ...item,
    ...('frontMatter' in item && { href: item.href || `${item.route.replace(/\/$/, '')}/` }),
    ...(item.children && { children: withExportLinks(item.children) }),
  }))
}

export default async function RootLayout({ children }) {
  return (
    <html lang="en" dir="ltr" suppressHydrationWarning>
      <Head color={{ hue: 85, saturation: 78, lightness: { light: 27, dark: 66 } }} backgroundColor={{ light: 'rgb(250, 249, 246)', dark: 'rgb(20, 24, 23)' }} />
      <body>
        <Layout
          navbar={<Navbar logo={<span className="wordmark"><span className="brand-emoji" aria-hidden="true">👨‍🔧</span>maintainer</span>} projectLink="https://github.com/octomation/maintainer" />}
          pageMap={withExportLinks(groupReleases(await getPageMap()))}
          docsRepositoryBase="https://github.com/octomation/maintainer/tree/main/docs"
          footer={<Footer><div className="site-footer"><span><b>maintainer</b> <span>Contribution assistant for open source projects.</span></span><span>MIT © {new Date().getFullYear()} <a href="https://github.com/octolab">OctoLab</a></span></div></Footer>}
          search={null}
          copyPageButton={false}
          feedback={{ content: 'Something unclear?', link: 'https://github.com/octomation/maintainer/issues/new' }}
          nextThemes={{ defaultTheme: 'system' }}
        >
          {children}
        </Layout>
      </body>
    </html>
  )
}
