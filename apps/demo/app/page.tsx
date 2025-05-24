"use client";

import { useEffect, useState, useRef } from "react";
import DataTable from "@/components/DataTable";
import Image from "next/image";

interface Row {
  id: string;
  organization_id: string;
  name: string;
  email: string;
  password: string;
  created_at: string;
  updated_at: string;
}

export default function Home() {
  const [primaryRows, setPrimaryRows] = useState<Row[] | null>(null);
  const [replicaRows, setReplicaRows] = useState<Row[] | null>(null);
  const [loading, setLoading] = useState(true);
  const [primaryEdits, setPrimaryEdits] = useState<Row[]>([]);
  const [saving, setSaving] = useState(false);
  const firstLoad = useRef(true);

  useEffect(() => {
    let isMounted = true;
    async function fetchData() {
      if (firstLoad.current) {
        setLoading(true);
      }
      const [primaryRes, replicaRes] = await Promise.all([
        fetch("/api/primary-table"),
        fetch("/api/replica-table"),
      ]);
      const [primaryData, replicaData] = await Promise.all([
        primaryRes.json(),
        replicaRes.json(),
      ]);
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

  const handlePrimaryEdit = (editedRows: Row[]) => {
    setPrimaryEdits(editedRows);
  };

  const handleSave = async () => {
    setSaving(true);
    await fetch("/api/primary-table", {
      method: "PATCH",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ rows: primaryEdits }),
    });
    setPrimaryEdits([]);
    setSaving(false);
  };

  return (
    <div className="min-h-screen flex flex-col bg-base-200 relative">
      <div className="absolute top-2 left-2 z-10">
        <Image src="/kasho-icon.png" alt="Kasho Icon" width={32} height={32} />
      </div>
      <div className="flex-1 border-b border-base-300 bg-base-200">
        <div
          className="max-w-6xl mx-auto w-full flex items-center gap-2"
          tabIndex={-1}
          onKeyDown={(e) => {
            if (primaryEdits.length > 0 && (e.key === 'Enter' || e.key === 'Return')) {
              handleSave();
            }
          }}
        >
          <span className="text-xl font-bold mb-2 text-primary">Primary</span>
          {primaryEdits.length > 0 && (
            <button
              className="btn btn-xs btn-success"
              onClick={handleSave}
              disabled={saving}
            >
              {saving ? "Saving..." : "Save"}
            </button>
          )}
        </div>
        <DataTable
          rows={primaryRows}
          loading={loading}
          editable={true}
          onEdit={handlePrimaryEdit}
          onSave={handleSave}
        />
      </div>
      <div className="flex-1 bg-base-300">
        <div className="max-w-6xl mx-auto w-full">
          <span className="text-xl font-bold mb-2 text-accent">Replica</span>
        </div>
        <DataTable rows={replicaRows} loading={loading} />
      </div>
    </div>
  );
}
