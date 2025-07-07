"use client";

import Image from "next/image";
import { useEffect, useState } from "react";
import { ObfuscatedEmail } from "../components/ObfuscatedEmail";

export default function Home() {
  const isDev = process.env.NODE_ENV === "development";
  const demoUrl = isDev ? "http://localhost:3001" : "https://demo.kasho.io";

  const [isDarkMode, setIsDarkMode] = useState(false);

  useEffect(() => {
    // Check initial theme preference
    const checkTheme = () => {
      const isDark = window.matchMedia("(prefers-color-scheme: dark)").matches;
      setIsDarkMode(isDark);
    };

    checkTheme();

    // Listen for theme changes
    const mediaQuery = window.matchMedia("(prefers-color-scheme: dark)");
    mediaQuery.addEventListener("change", checkTheme);

    return () => mediaQuery.removeEventListener("change", checkTheme);
  }, []);

  return (
    <div className="h-screen flex flex-col items-center justify-start px-4 overflow-hidden relative bg-base-100">
      <div className="mt-16 w-full flex flex-col items-center">
        <Image
          src={isDarkMode ? "/kasho-wordmark-dark.png" : "/kasho-wordmark-light.png"}
          alt="Kasho Wordmark"
          width={2400}
          height={1200}
          className="w-full max-w-xs sm:max-w-md md:max-w-lg lg:max-w-2xl h-auto"
          priority
        />
        <div className="mt-4 sm:mt-6 flex flex-col items-center w-full">
          <h2 className="text-base-content text-lg sm:text-xl md:text-2xl font-bold font-mono text-center max-w-md sm:max-w-lg">
            Anonymized, live replicas on demand for development, testing and staging.
          </h2>
          <h3 className="text-base-content text-lg sm:text-xl md:text-2xl font-bold font-mono text-center mt-4">
            COMING SOON.
          </h3>
        </div>
        <div className="mt-8 w-full max-w-lg">
          <div className="alert alert-info shadow-lg">
            <svg
              xmlns="http://www.w3.org/2000/svg"
              fill="none"
              viewBox="0 0 24 24"
              className="stroke-current shrink-0 w-6 h-6"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth="2"
                d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
              ></path>
            </svg>
            <div>
              <h3 className="font-bold">Seeking Design Partners</h3>
              <div className="text-sm">
                We&apos;re looking for early adopters to help shape Kasho. Interested?{" "}
                <span className="font-semibold">
                  Contact us at <ObfuscatedEmail user="hello" domain="kasho.io" className="link link-hover" />
                </span>
              </div>
            </div>
          </div>
        </div>
      </div>
      <a
        href={demoUrl}
        target="_blank"
        rel="noopener noreferrer"
        className="absolute left-1/2 -translate-x-1/2 text-sm text-primary hover:text-primary-focus font-semibold underline"
        style={{ bottom: "1.5rem", paddingBottom: "env(safe-area-inset-bottom)" }}
      >
        Try the live demo &rarr;
      </a>
    </div>
  );
}
// Test homepage change
