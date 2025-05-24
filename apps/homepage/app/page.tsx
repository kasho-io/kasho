import Image from "next/image";

export default function Home() {
  return (
    <div className="min-h-screen flex flex-col items-center justify-center" style={{ backgroundColor: '#101110' }}>
      <img
        src="/kasho-wordmark.png"
        alt="Kasho Wordmark"
        className="max-w-2xl h-auto"
      />
      <div className="mt-8 flex flex-col items-center">
        <h2 className="text-white text-2xl font-bold font-mono text-center">
          Anonymized, live replicas on demand for development, testing and staging.
        </h2>
        <br/>
        <h3 className="text-white text-2xl font-bold font-mono text-center">COMING SOON.</h3>
      </div>
    </div>
  );
}
