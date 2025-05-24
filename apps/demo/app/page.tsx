"use client";

import { useEffect, useState, useRef } from "react";
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
