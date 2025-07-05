import { Footer, Layout, Navbar } from 'nextra-theme-docs'
import { Banner, Head } from 'nextra/components'
import { getPageMap } from 'nextra/page-map'
import { Inter, JetBrains_Mono } from 'next/font/google'
import Image from 'next/image'
import 'nextra-theme-docs/style.css'
import './globals.css'

const inter = Inter({
  variable: '--font-inter',
  subsets: ['latin'],
  display: 'swap',
})

const jetbrainsMono = JetBrains_Mono({
  variable: '--font-jetbrains-mono',
  subsets: ['latin'],
  display: 'swap',
})
 
export const metadata = {
  title: 'Kasho Documentation',
  description: 'Documentation for Kasho - Anonymized, live replicas on demand',
}
 
const banner = null // Remove the banner for now
const navbar = (
  <Navbar
    logo={
      <>
        <img 
          src="/kasho-wordmark-light.png" 
          alt="Kasho" 
          className="nextra-logo-light"
        />
        <img 
          src="/kasho-wordmark-dark.png" 
          alt="Kasho" 
          className="nextra-logo-dark"
        />
      </>
    }
    // ... Your additional navbar options
  />
)
const footer = <Footer>© {new Date().getFullYear()} Kasho</Footer>
 
export default async function RootLayout({ children }) {
  return (
    <html
      // Not required, but good for SEO
      lang="en"
      // Required to be set
      dir="ltr"
      // Suggested by `next-themes` package https://github.com/pacocoursey/next-themes#with-app
      suppressHydrationWarning
    >
      <Head
      // ... Your additional head options
      >
        {/* Your additional tags should be passed as `children` of `<Head>` element */}
      </Head>
      <body className={`${inter.variable} ${jetbrainsMono.variable}`}>
        <Layout
          banner={banner}
          navbar={navbar}
          pageMap={await getPageMap()}
          docsRepositoryBase="https://github.com/kasho-org/kasho/tree/main/apps/docs"
          footer={footer}
          // ... Your additional layout options
        >
          {children}
        </Layout>
      </body>
    </html>
  )
}