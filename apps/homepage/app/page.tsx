import Image from "next/image";

export default function Home() {
  return (
    <div className="min-h-screen flex flex-col items-center justify-center px-4" style={{ backgroundColor: '#101110' }}>
      <Image
        src="/kasho-wordmark.png"
        alt="Kasho Wordmark"
        width={2400}
        height={1200}
        className="w-full max-w-xs sm:max-w-md md:max-w-lg lg:max-w-2xl h-auto"
        priority
      />
      <div className="mt-6 sm:mt-8 flex flex-col items-center w-full">
        <h2 className="text-white text-lg sm:text-xl md:text-2xl font-bold font-mono text-center max-w-md sm:max-w-lg">
          Anonymized, live replicas on demand for development, testing and staging.
        </h2>
        <h3 className="text-white text-lg sm:text-xl md:text-2xl font-bold font-mono text-center mt-4">COMING SOON.</h3>
      </div>
    </div>
  );
}
