"use client";

import { useEffect, useState } from "react";

interface Row {
  id: number;
  organization_id: number;
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
      {/* Top Pane */}
      <div className="flex-1 flex flex-col items-center justify-center border-b border-base-300 p-8">
        <h2 className="text-xl font-bold mb-4">Primary</h2>
        <div className="w-full max-w-3xl overflow-x-auto">
          <table className="table table-zebra w-full font-mono min-w-[900px]">
            <thead>
              <tr>
                <th>ID</th>
                <th>Organization ID</th>
                <th>Name</th>
                <th>Email</th>
                <th>Password</th>
                <th>Created At</th>
                <th>Updated At</th>
              </tr>
            </thead>
            <tbody>
              {loading && (
                <tr>
                  <td colSpan={7} className="text-center">Loading...</td>
                </tr>
              )}
              {!loading && Array.isArray(primaryRows) && primaryRows.map((row) => (
                <tr key={row.id}>
                  <td>{row.id}</td>
                  <td>{row.organization_id}</td>
                  <td>{row.name}</td>
                  <td>{row.email}</td>
                  <td>{row.password}</td>
                  <td>{row.created_at}</td>
                  <td>{row.updated_at}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
      {/* Bottom Pane */}
      <div className="flex-1 flex flex-col items-center justify-center p-8">
        <h2 className="text-xl font-bold mb-4">Replica</h2>
        <div className="w-full max-w-3xl overflow-x-auto">
          <table className="table table-zebra w-full font-mono min-w-[900px]">
            <thead>
              <tr>
                <th>ID</th>
                <th>Organization ID</th>
                <th>Name</th>
                <th>Email</th>
                <th>Password</th>
                <th>Created At</th>
                <th>Updated At</th>
              </tr>
            </thead>
            <tbody>
              {loading && (
                <tr>
                  <td colSpan={7} className="text-center">Loading...</td>
                </tr>
              )}
              {!loading && Array.isArray(replicaRows) && replicaRows.map((row) => (
                <tr key={row.id}>
                  <td>{row.id}</td>
                  <td>{row.organization_id}</td>
                  <td>{row.name}</td>
                  <td>{row.email}</td>
                  <td>{row.password}</td>
                  <td>{row.created_at}</td>
                  <td>{row.updated_at}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}
