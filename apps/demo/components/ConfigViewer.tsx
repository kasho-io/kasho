import { useState, useEffect } from "react";

// Simple YAML syntax highlighter
function highlightYaml(yamlText: string): string {
  return (
    yamlText
      // Keys (before colon)
      .replace(/^(\s*)([\w\-_]+)(\s*:)/gm, '$1<span class="text-blue-400 font-semibold">$2</span>$3')
      // String values (quoted)
      .replace(/:\s*(['"]).+?\1/g, (match) => match.replace(/(['"]).+?\1/, '<span class="text-green-400">$&</span>'))
      // Comments
      .replace(/(#.*)$/gm, '<span class="text-gray-500 italic">$1</span>')
      // Version values
      .replace(/:\s*(v\d+)/g, ': <span class="text-purple-400">$1</span>')
      // Transform types
      .replace(
        /:\s*(FakeName|Template|PasswordArgon2id|PasswordBcrypt|PasswordScrypt|PasswordPBKDF2)/g,
        ': <span class="text-orange-400 font-medium">$1</span>',
      )
      // Template syntax
      .replace(/({{[^}]+}})/g, '<span class="text-cyan-400">$1</span>')
  );
}

interface ConfigData {
  content: string;
  source: string;
}

interface ConfigViewerProps {
  className?: string;
}

export default function ConfigViewer({ className = "" }: ConfigViewerProps) {
  const [config, setConfig] = useState<ConfigData | null>(null);
  const [isExpanded, setIsExpanded] = useState(false);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    async function fetchConfig() {
      try {
        const response = await fetch("/api/config");
        const data = await response.json();
        setConfig(data);
      } catch (err) {
        console.error("Failed to fetch config:", err);
      } finally {
        setLoading(false);
      }
    }
    fetchConfig();
  }, []);

  return (
    <div className={`border border-base-300 rounded-lg bg-base-100 ${className}`}>
      <button
        onClick={() => setIsExpanded(!isExpanded)}
        className="w-full p-3 text-left flex items-center justify-between hover:bg-base-200 rounded-t-lg"
      >
        <div className="flex items-center gap-2">
          <span className="text-sm font-semibold">
            <span className="font-mono">transforms.yml</span> Configuration
          </span>
        </div>
        <div className={`transform transition-transform ${isExpanded ? "rotate-180" : ""}`}>
          <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
          </svg>
        </div>
      </button>

      {isExpanded && (
        <div className="border-t border-base-300">
          {loading ? (
            <div className="p-4 text-center text-sm opacity-60">Loading config...</div>
          ) : config ? (
            <pre
              className="p-4 text-xs font-mono bg-base-200 overflow-x-auto rounded-b-lg whitespace-pre"
              dangerouslySetInnerHTML={{ __html: highlightYaml(config.content) }}
            />
          ) : (
            <div className="p-4 text-center text-sm text-error">Failed to load config</div>
          )}
        </div>
      )}
    </div>
  );
}
