import { Footer, Layout, Navbar } from 'nextra-theme-docs'
import { Head } from 'nextra/components'
import { getPageMap } from 'nextra/page-map'
import 'nextra-theme-docs/style.css'
import './globals.css'
import { siteUrl } from '../site.mjs'

export const metadata = {
  title: { default: 'maintainer — Small tools for open source upkeep', template: '%s · maintainer' },
  description: 'See your open source contribution pace and plan your next session from the terminal.',
  metadataBase: new URL(siteUrl),
  icons: { icon: `${process.env.BASE_PATH || ''}/icon.svg` },
}

function withExportLinks(items) {
  return items.map(item => ({ ...item,
    ...('frontMatter' in item && { href: item.href || `${item.route.replace(/\/$/, '')}/` }),
    ...(item.children && { children: withExportLinks(item.children) }),
  }))
}

export default async function RootLayout({ children }) {
  return <html lang="en" dir="ltr" suppressHydrationWarning>
    <Head color={{ hue: 211, saturation: 70 }} />
    <body><Layout
      navbar={<Navbar logo={<b className="wordmark">👨‍🔧 maintainer</b>} projectLink="https://github.com/octomation/maintainer" />}
      pageMap={withExportLinks(await getPageMap())}
      docsRepositoryBase="https://github.com/octomation/maintainer/tree/main/docs"
      footer={<Footer>MIT © {new Date().getFullYear()} OctoLab</Footer>}
      search={null}
      copyPageButton={false}
      feedback={{ content: 'Something unclear?', link: 'https://github.com/octomation/maintainer/issues/new' }}
      nextThemes={{ defaultTheme: 'system' }}
    >{children}</Layout></body>
  </html>
}
