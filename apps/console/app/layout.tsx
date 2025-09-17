import type { Metadata } from "next";
import { Inter, JetBrains_Mono } from "next/font/google";
import Navigation from "@/components/Navigation";
import ClientLayout from "@/components/ClientLayout";
import { services } from "@/lib/services";
import "./globals.css";

const inter = Inter({
  variable: "--font-inter",
  subsets: ["latin"],
  display: "swap",
});

const jetbrainsMono = JetBrains_Mono({
  variable: "--font-jetbrains-mono",
  subsets: ["latin"],
  display: "swap",
});

export const metadata: Metadata = {
  title: "Kasho",
  description: "Anonymized, live replicas on demand for development, testing and staging.",
};

export default async function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  // Check if user has an organization
  let hasOrganization = false;
  try {
    const session = await services.workos.withAuth();
    hasOrganization = !!session.organizationId;
  } catch {
    // User not authenticated or error - default to false
  }

  return (
    <html lang="en">
      <body className={`${inter.variable} ${jetbrainsMono.variable} font-sans antialiased`}>
        <Navigation />
        <ClientLayout hasOrganization={hasOrganization}>
          <main>{children}</main>
        </ClientLayout>
      </body>
    </html>
  );
}
