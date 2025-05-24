interface Row {
  id: string;
  organization_id: string;
  name: string;
  email: string;
  password: string;
  created_at: string;
  updated_at: string;
}

interface DataTableProps {
  title: string;
  rows: Row[] | null;
  loading: boolean;
}

function lastSegment(uuid: string) {
  return uuid.split('-').pop() || uuid;
}

function last8(str: string) {
  return str.length > 8 ? str.slice(-8) : str;
}

function formatDate(dateStr: string) {
  const date = new Date(dateStr);
  if (isNaN(date.getTime())) return dateStr;
  return date.toLocaleString(undefined, {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  });
}

export default function DataTable({ title, rows, loading }: DataTableProps) {
  return (
    <div className="flex-1 flex flex-col items-center justify-center p-2">
      <h2 className={`text-xl font-bold mb-2 ${title === 'Primary' ? 'text-primary' : title === 'Replica' ? 'text-accent' : ''}`}>{title}</h2>
      <div className="w-full max-w-6xl overflow-x-auto">
        <table className="table table-zebra w-full font-mono min-w-[1200px] bg-base-100">
          <thead>
            <tr>
              <th className="p-2 whitespace-nowrap">ID</th>
              <th className="p-2 whitespace-nowrap">Organization ID</th>
              <th className="p-2 whitespace-nowrap">Name</th>
              <th className="p-2 whitespace-nowrap">Email</th>
              <th className="p-2 whitespace-nowrap">Password</th>
              <th className="p-2 whitespace-nowrap">Created At</th>
              <th className="p-2 whitespace-nowrap">Updated At</th>
            </tr>
          </thead>
          <tbody>
            {loading && (
              <tr>
                <td colSpan={7} className="text-center p-2 whitespace-nowrap">Loading...</td>
              </tr>
            )}
            {!loading && Array.isArray(rows) && rows.map((row) => (
              <tr key={row.id}>
                <td className="p-2 whitespace-nowrap" title={row.id}>{'…' + lastSegment(row.id)}</td>
                <td className="p-2 whitespace-nowrap" title={row.organization_id}>{'…' + lastSegment(row.organization_id)}</td>
                <td className="p-2 whitespace-nowrap">{row.name}</td>
                <td className="p-2 whitespace-nowrap">{row.email}</td>
                <td className="p-2 whitespace-nowrap" title={row.password}>{'…' + last8(row.password)}</td>
                <td className="p-2 whitespace-nowrap" title={row.created_at}>{formatDate(row.created_at)}</td>
                <td className="p-2 whitespace-nowrap" title={row.updated_at}>{formatDate(row.updated_at)}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
} 