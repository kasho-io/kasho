import { useMDXComponents as getThemeComponents } from 'nextra-theme-docs'
import type { MDXComponents } from 'nextra/mdx-components'
import { ObfuscatedEmail } from './components/ObfuscatedEmail'

// Get the default MDX components
const themeComponents = getThemeComponents({})

// Merge components
export function useMDXComponents(components: MDXComponents = {}): MDXComponents {
  return {
    ...themeComponents,
    ObfuscatedEmail,
    ...components
  }
}