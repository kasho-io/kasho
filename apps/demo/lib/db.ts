import { createPool, DatabasePool, detectDatabaseType, DatabaseType } from "./db-factory";

let primaryPool: DatabasePool | null = null;
let replicaPool: DatabasePool | null = null;

function getPrimaryPool(): DatabasePool {
  if (!primaryPool) {
    const connectionString = process.env.PRIMARY_DATABASE_URL;
    if (!connectionString) {
      throw new Error("PRIMARY_DATABASE_URL is not set");
    }
    primaryPool = createPool(connectionString);
  }
  return primaryPool;
}

function getReplicaPool(): DatabasePool {
  if (!replicaPool) {
    const connectionString = process.env.REPLICA_DATABASE_URL;
    if (!connectionString) {
      throw new Error("REPLICA_DATABASE_URL is not set");
    }
    replicaPool = createPool(connectionString);
  }
  return replicaPool;
}

export async function queryPrimary(sql: string, params?: unknown[]) {
  const result = await getPrimaryPool().query(sql, params);
  return result.rows;
}

export async function queryReplica(sql: string, params?: unknown[]) {
  const result = await getReplicaPool().query(sql, params);
  return result.rows;
}

// Export database type detection for UI purposes
export function getPrimaryDatabaseType(): DatabaseType {
  return detectDatabaseType(process.env.PRIMARY_DATABASE_URL || "");
}

export function getReplicaDatabaseType(): DatabaseType {
  return detectDatabaseType(process.env.REPLICA_DATABASE_URL || "");
}
