import { generateStaticParamsFor, importPage } from 'nextra/pages'
import { useMDXComponents } from '../../mdx-components'
import { siteUrl } from '../../site.mjs'

export const generateStaticParams = generateStaticParamsFor('mdxPath')

export async function generateMetadata({ params }) {
  const { mdxPath } = await params
  const page = await importPage(mdxPath)
  const { title, description } = page.metadata
  const path = mdxPath?.length ? `${mdxPath.join('/')}/` : ''
  const url = `${siteUrl}${path}`
  const images = path ? [] : [{
    url: `${siteUrl}og.png`,
    width: 1733,
    height: 908,
    alt: 'maintainer — Small tools for open source upkeep',
  }]
  return {
    ...page.metadata,
    alternates: { canonical: url },
    openGraph: { title, description, url, siteName: 'maintainer', type: 'website', images },
    twitter: { card: path ? 'summary' : 'summary_large_image', title, description, images },
  }
}

const Wrapper = useMDXComponents().wrapper

export default async function Page({ params }) {
  const resolvedParams = await params
  const { default: Content, toc, metadata, sourceCode } = await importPage(resolvedParams.mdxPath)

  return (
    <Wrapper toc={toc} metadata={metadata} sourceCode={sourceCode}>
      <Content params={resolvedParams} />
    </Wrapper>
  )
}
