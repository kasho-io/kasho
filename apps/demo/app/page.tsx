"use client";

import { useEffect, useState, useRef, useCallback } from "react";
import DataTable from "@/components/DataTable";
import ConfigViewer from "@/components/ConfigViewer";
import Image from "next/image";

interface Row {
  id: string;
  organization_id: string;
  name: string;
  email: string;
  password: string;
  updated_at: string;
}

export default function Home() {
  const [primaryRows, setPrimaryRows] = useState<Row[] | null>(null);
  const [replicaRows, setReplicaRows] = useState<Row[] | null>(null);
  const [loading, setLoading] = useState(true);
  const [primaryEdits, setPrimaryEdits] = useState<Row[]>([]);
  const [saving, setSaving] = useState(false);
  const [isDarkMode, setIsDarkMode] = useState(false);
  const [saveKey, setSaveKey] = useState(0);
  const [instructionsExpanded, setInstructionsExpanded] = useState(false);
  const firstLoad = useRef(true);

  useEffect(() => {
    let isMounted = true;
    async function fetchData() {
      if (firstLoad.current) {
        setLoading(true);
      }
      const [primaryRes, replicaRes] = await Promise.all([fetch("/api/primary-table"), fetch("/api/replica-table")]);
      const [primaryData, replicaData] = await Promise.all([primaryRes.json(), replicaRes.json()]);
      if (!isMounted) return;
      setPrimaryRows(primaryData);
      setReplicaRows(replicaData);
      if (firstLoad.current) {
        setLoading(false);
        firstLoad.current = false;
      }
    }
    fetchData();
    const interval = setInterval(fetchData, 3000);
    return () => {
      isMounted = false;
      clearInterval(interval);
    };
  }, []);

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

  const handlePrimaryEdit = useCallback((editedRows: Row[]) => {
    setPrimaryEdits(editedRows);
  }, []);

  const handleSave = async () => {
    setSaving(true);
    await fetch("/api/primary-table", {
      method: "PATCH",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ rows: primaryEdits }),
    });
    setPrimaryEdits([]);
    setSaveKey((prev) => prev + 1); // Force DataTable to reset

    // Immediately fetch updated data
    const [primaryRes, replicaRes] = await Promise.all([fetch("/api/primary-table"), fetch("/api/replica-table")]);
    const [primaryData, replicaData] = await Promise.all([primaryRes.json(), replicaRes.json()]);
    setPrimaryRows(primaryData);
    setReplicaRows(replicaData);

    setSaving(false);
  };

  return (
    <div className="min-h-screen flex flex-col bg-base-200 relative pt-14 sm:pt-0">
      {/* Mobile: centered, above everything */}
      <div className="fixed top-0 left-0 w-full flex justify-center z-20 sm:hidden bg-base-200 pt-2 pb-2">
        <Image
          src={isDarkMode ? "/kasho-icon-dark.png" : "/kasho-icon-light.png"}
          alt="Kasho Icon"
          width={32}
          height={32}
        />
      </div>
      {/* Desktop: top-left */}
      <div className="absolute top-2 left-2 z-10 hidden sm:block">
        <Image
          src={isDarkMode ? "/kasho-icon-dark.png" : "/kasho-icon-light.png"}
          alt="Kasho Icon"
          width={32}
          height={32}
        />
      </div>

      {/* Header Section */}
      <div className="bg-base-100 border-b border-base-300 p-4">
        <div className="max-w-6xl mx-auto space-y-4">
          <div>
            <h1 className="text-2xl font-bold mb-2">Kasho Live Demo</h1>
            <p className="text-sm opacity-70 mb-3">
              Real-time database replication with data transformation. Watch changes propagate from primary to replica
              with live transforms.
            </p>
          </div>
          <ConfigViewer />
        </div>
      </div>
      <div className="flex-1 border-b border-base-300 bg-base-200">
        <div className="max-w-6xl mx-auto w-full p-2">
          <div
            className="flex items-center gap-2 mb-2"
            tabIndex={-1}
            onKeyDown={(e) => {
              if (primaryEdits.length > 0 && (e.key === "Enter" || e.key === "Return")) {
                handleSave();
              }
            }}
          >
            <span className="text-xl font-bold text-primary">
              Primary{" "}
              <span className="text-sm font-normal opacity-70 font-mono">(primary_db@postgres-primary:5432)</span>
            </span>
            {primaryEdits.length > 0 && (
              <button className="btn btn-xs btn-success" onClick={handleSave} disabled={saving}>
                {saving ? "Saving..." : "Save"}
              </button>
            )}
          </div>
          <div className="mb-3 bg-info/10 border border-info/20 rounded overflow-hidden">
            <button
              onClick={() => setInstructionsExpanded(!instructionsExpanded)}
              className="w-full p-2 text-left flex items-center justify-between hover:bg-info/20 transition-colors"
            >
              <span className="text-xs font-semibold text-info">💡 Try editing!</span>
              <div className={`transform transition-transform ${instructionsExpanded ? "rotate-180" : ""}`}>
                <svg className="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
                </svg>
              </div>
            </button>

            {instructionsExpanded && (
              <div className="border-t border-info/20 p-3 text-xs opacity-70 space-y-2">
                <div>
                  <span className="font-semibold text-info">name:</span> Will use a deterministic fake name (based on
                  your input) generated by Kasho
                </div>
                <div>
                  <span className="font-semibold text-info">email:</span> Edits will be ignored during transformation -
                  email will be generated based on the transformed name using template:{" "}
                  <code className="font-mono bg-black/10 px-1 rounded">{"{{.name | slugify}}@example.com"}</code>
                </div>
                <div>
                  <span className="font-semibold text-info">password:</span> Edits will be ignored during transformation
                  - password will be encrypted &apos;staging&apos; with Argon2id + salt for all users in replica
                </div>
              </div>
            )}
          </div>
        </div>
        <DataTable
          key={saveKey}
          rows={primaryRows}
          loading={loading}
          editable={true}
          onEdit={handlePrimaryEdit}
          onSave={handleSave}
        />
      </div>
      <div className="flex-1 bg-base-300">
        <div className="max-w-6xl mx-auto w-full">
          <span className="text-xl font-bold mb-2 text-accent">
            Replica <span className="text-sm font-normal opacity-70 font-mono">(replica_db@postgres-replica:5432)</span>
          </span>
        </div>
        <DataTable rows={replicaRows} loading={loading} />
      </div>
    </div>
  );
}
// Test comment
// Another test comment
// Test demo change
// Another test demo change
