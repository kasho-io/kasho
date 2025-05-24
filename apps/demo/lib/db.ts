import { Pool } from 'pg';

const primaryPool = new Pool({
  connectionString: process.env.PRIMARY_DATABASE_URL,
});

const replicaPool = new Pool({
  connectionString: process.env.REPLICA_DATABASE_URL,
});

export async function queryPrimary(sql: string, params?: any[]) {
  const result = await primaryPool.query(sql, params);
  return result.rows;
}

export async function queryReplica(sql: string, params?: any[]) {
  const result = await replicaPool.query(sql, params);
  return result.rows;
} 