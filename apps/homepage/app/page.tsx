import Image from "next/image";

export default function Home() {
  const isDev = process.env.NODE_ENV === "development";
  const demoUrl = isDev ? "http://localhost:4000" : "https://demo.kasho.io";

  return (
    <div className="h-screen flex flex-col items-center justify-start px-4 overflow-hidden relative" style={{ backgroundColor: '#101110' }}>
      <div className="mt-16 w-full flex flex-col items-center">
        <Image
          src="/kasho-wordmark.png"
          alt="Kasho Wordmark"
          width={2400}
          height={1200}
          className="w-full max-w-xs sm:max-w-md md:max-w-lg lg:max-w-2xl h-auto"
          priority
        />
        <div className="mt-4 sm:mt-6 flex flex-col items-center w-full">
          <h2 className="text-white text-lg sm:text-xl md:text-2xl font-bold font-mono text-center max-w-md sm:max-w-lg">
            Anonymized, live replicas on demand for development, testing and staging.
          </h2>
          <h3 className="text-white text-lg sm:text-xl md:text-2xl font-bold font-mono text-center mt-4">COMING SOON.</h3>
        </div>
      </div>
      <a
        href={demoUrl}
        target="_blank"
        rel="noopener noreferrer"
        className="absolute bottom-6 left-1/2 -translate-x-1/2 text-sm text-blue-400 hover:text-blue-200 font-semibold underline"
      >
        Try the live demo &rarr;
      </a>
    </div>
  );
}
