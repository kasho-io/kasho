"use client";

import Image from "next/image";
import { useEffect, useState } from "react";

export default function Home() {
  const isDev = process.env.NODE_ENV === "development";
  const demoPort = process.env.APP_DEMO_PORT ?? "3001";
  const docsPort = process.env.APP_DOCS_PORT ?? "3002";
  const demoUrl = isDev ? `http://localhost:${demoPort}` : "https://demo.kasho.io";
  const docsUrl = isDev ? `http://localhost:${docsPort}` : "https://docs.kasho.io";
  const githubUrl = "https://github.com/kasho-io/kasho";

  const [isDarkMode, setIsDarkMode] = useState(false);

  useEffect(() => {
    const checkTheme = () => {
      const isDark = window.matchMedia("(prefers-color-scheme: dark)").matches;
      setIsDarkMode(isDark);
    };

    checkTheme();

    const mediaQuery = window.matchMedia("(prefers-color-scheme: dark)");
    mediaQuery.addEventListener("change", checkTheme);

    return () => mediaQuery.removeEventListener("change", checkTheme);
  }, []);

  return (
    <div className="min-h-screen flex flex-col items-center justify-center px-4 bg-base-100">
      <div className="w-full flex flex-col items-center">
        <Image
          src={isDarkMode ? "/kasho-wordmark-dark.png" : "/kasho-wordmark-light.png"}
          alt="Kasho Wordmark"
          width={2400}
          height={1200}
          className="w-full max-w-xs sm:max-w-md md:max-w-lg lg:max-w-2xl h-auto"
          priority
        />
        <p className="text-base-content text-lg sm:text-xl md:text-2xl font-bold font-mono text-center max-w-md sm:max-w-lg mt-4 sm:mt-6">
          Anonymized, live replicas on demand for development, testing and staging.
        </p>
        <div className="flex flex-wrap justify-center gap-4 mt-8">
          <a href={docsUrl} target="_blank" rel="noopener noreferrer" className="btn btn-outline gap-2">
            <svg
              xmlns="http://www.w3.org/2000/svg"
              fill="none"
              viewBox="0 0 24 24"
              strokeWidth={1.5}
              stroke="currentColor"
              className="w-5 h-5"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                d="M12 6.042A8.967 8.967 0 0 0 6 3.75c-1.052 0-2.062.18-3 .512v14.25A8.987 8.987 0 0 1 6 18c2.305 0 4.408.867 6 2.292m0-14.25a8.966 8.966 0 0 1 6-2.292c1.052 0 2.062.18 3 .512v14.25A8.987 8.987 0 0 0 18 18a8.967 8.967 0 0 0-6 2.292m0-14.25v14.25"
              />
            </svg>
            Docs
          </a>
          <a href={demoUrl} target="_blank" rel="noopener noreferrer" className="btn btn-outline gap-2">
            <svg
              xmlns="http://www.w3.org/2000/svg"
              fill="none"
              viewBox="0 0 24 24"
              strokeWidth={1.5}
              stroke="currentColor"
              className="w-5 h-5"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                d="M5.25 5.653c0-.856.917-1.398 1.667-.986l11.54 6.347a1.125 1.125 0 0 1 0 1.972l-11.54 6.347a1.125 1.125 0 0 1-1.667-.986V5.653Z"
              />
            </svg>
            Demo
          </a>
          <a href={githubUrl} target="_blank" rel="noopener noreferrer" className="btn btn-outline gap-2">
            <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor" className="w-5 h-5">
              <path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z" />
            </svg>
            GitHub
          </a>
        </div>
      </div>
    </div>
  );
}
