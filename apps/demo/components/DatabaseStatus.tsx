interface DatabaseStatusProps {
  className?: string;
}

export default function DatabaseStatus({ className = "" }: DatabaseStatusProps) {
  return (
    <div className={`flex flex-col sm:flex-row gap-4 ${className}`}>
      <div className="flex items-center gap-2">
        <div className="w-3 h-3 bg-success rounded-full"></div>
        <span className="text-sm font-mono">
          <span className="font-semibold text-primary">Primary DB:</span> primary_db (postgres-primary:5432)
        </span>
      </div>
      <div className="flex items-center gap-2">
        <div className="w-3 h-3 bg-success rounded-full"></div>
        <span className="text-sm font-mono">
          <span className="font-semibold text-accent">Replica DB:</span> replica_db (postgres-replica:5432)
        </span>
      </div>
    </div>
  );
}
