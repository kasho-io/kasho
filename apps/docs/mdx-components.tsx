import { useMDXComponents as getThemeComponents } from 'nextra-theme-docs'
import type { MDXComponents } from 'nextra/mdx-components'

// Get the default MDX components
const themeComponents = getThemeComponents({})

// Merge components
export function useMDXComponents(components: MDXComponents = {}): MDXComponents {
  return {
    ...themeComponents,
    ...components
  }
}