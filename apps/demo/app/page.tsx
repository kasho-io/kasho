"use client";

import { useEffect, useState } from "react";
import DataTable from "@/components/DataTable";

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

  useEffect(() => {
    async function fetchData() {
      setLoading(true);
      const [primaryRes, replicaRes] = await Promise.all([
        fetch("/api/primary-table"),
        fetch("/api/replica-table"),
      ]);
      const [primaryData, replicaData] = await Promise.all([
        primaryRes.json(),
        replicaRes.json(),
      ]);
      console.log("Primary API response:", primaryData);
      console.log("Replica API response:", replicaData);
      setPrimaryRows(primaryData);
      setReplicaRows(replicaData);
      setLoading(false);
    }
    fetchData();
  }, []);

  return (
    <div className="min-h-screen flex flex-col bg-base-200">
      <div className="flex-1 border-b border-base-300 bg-base-200">
        <DataTable title="Primary" rows={primaryRows} loading={loading} />
      </div>
      <div className="flex-1 bg-base-300">
        <DataTable title="Replica" rows={replicaRows} loading={loading} />
      </div>
    </div>
  );
}
