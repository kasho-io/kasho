import { NextResponse } from "next/server";
import { detectDatabaseType, DatabaseType } from "@/lib/db-factory";

interface DatabaseInfo {
  type: DatabaseType;
  host: string;
  port: string;
  database: string;
}

function parseConnectionString(connectionString: string): DatabaseInfo {
  const dbType = detectDatabaseType(connectionString);

  try {
    const url = new URL(connectionString);
    return {
      type: dbType,
      host: url.hostname,
      port: url.port || (dbType === "mysql" ? "3306" : "5432"),
      database: url.pathname.slice(1), // Remove leading /
    };
  } catch {
    // Return defaults if parsing fails
    return {
      type: dbType,
      host: "localhost",
      port: dbType === "mysql" ? "3306" : "5432",
      database: "unknown",
    };
  }
}

export async function GET() {
  const primaryUrl = process.env.PRIMARY_DATABASE_URL || "";
  const replicaUrl = process.env.REPLICA_DATABASE_URL || "";

  return NextResponse.json({
    primary: parseConnectionString(primaryUrl),
    replica: parseConnectionString(replicaUrl),
  });
}
