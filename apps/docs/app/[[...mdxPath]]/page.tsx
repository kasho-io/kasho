import { generateStaticParamsFor, importPage } from 'nextra/pages'
import { useMDXComponents as getMDXComponents } from '../../mdx-components'
import type { Metadata } from 'next'
import type { Heading } from 'nextra'

interface PageProps {
  params: Promise<{
    mdxPath?: string[]
  }>
  searchParams: Promise<{
    [key: string]: string | string[] | undefined
  }>
}

// Since importPage returns Promise<any>, we need to define the result type
interface PageResult {
  default: React.ComponentType<{
    params: { mdxPath?: string[] }
    searchParams: Promise<{
      [key: string]: string | string[] | undefined
    }>
  }>
  toc: Heading[]
  metadata: Metadata
}

export const generateStaticParams = generateStaticParamsFor('mdxPath')

export async function generateMetadata({ params }: PageProps): Promise<Metadata> {
  const resolvedParams = await params
  const { metadata } = await importPage(resolvedParams.mdxPath)
  return metadata
}

const Wrapper = getMDXComponents(null).wrapper as React.ComponentType<{
  toc: Heading[]
  metadata: Metadata
  children: React.ReactNode
}>

export default async function Page(props: PageProps) {
  const params = await props.params
  const result = await importPage(params.mdxPath) as PageResult
  const { default: MDXContent, toc, metadata } = result
  return (
    <Wrapper toc={toc} metadata={metadata}>
      <MDXContent params={params} searchParams={props.searchParams} />
    </Wrapper>
  )
}