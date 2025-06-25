"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";

const navigation = [
  {
    title: "Getting Started",
    items: [
      { title: "Introduction", href: "/" },
      { title: "Quick Start", href: "/quick-start" },
    ],
  },
  {
    title: "Installation & Setup",
    items: [
      { title: "Database Setup", href: "/installation/database-setup" },
      { title: "Configuration", href: "/installation/configuration" },
      { title: "Bootstrap Process", href: "/installation/bootstrap" },
    ],
  },
  {
    title: "Configuration Reference",
    items: [
      { title: "Transform Configuration", href: "/configuration/transforms" },
    ],
  },
];

export function Navigation() {
  const pathname = usePathname();

  return (
    <aside className="w-64 min-h-screen bg-base-200 p-4">
      <div className="mb-8">
        <Link href="/" className="text-2xl font-bold">
          Kasho Docs
        </Link>
      </div>
      
      <nav className="space-y-8">
        {navigation.map((section) => (
          <div key={section.title}>
            <h3 className="font-semibold text-xs uppercase tracking-wider text-base-content/60 mb-3">
              {section.title}
            </h3>
            <ul className="space-y-0.5">
              {section.items.map((item) => {
                const isActive = pathname === item.href;
                return (
                  <li key={item.href}>
                    <Link
                      href={item.href}
                      className={`
                        block px-3 py-1.5 rounded-md text-sm transition-colors
                        ${
                          isActive
                            ? "bg-primary text-primary-content font-medium"
                            : "text-base-content/80 hover:text-base-content hover:bg-base-300/50"
                        }
                      `}
                    >
                      {item.title}
                    </Link>
                  </li>
                );
              })}
            </ul>
          </div>
        ))}
      </nav>
    </aside>
  );
}