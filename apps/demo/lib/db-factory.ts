import { Pool as PgPool } from "pg";
import mysql from "mysql2/promise";

export interface DatabasePool {
  query(sql: string, params?: unknown[]): Promise<{ rows: Record<string, unknown>[] }>;
  end(): Promise<void>;
}

export type DatabaseType = "postgresql" | "mysql";

// Detect database type from connection string
export function detectDatabaseType(connectionString: string): DatabaseType {
  if (connectionString.startsWith("mysql://")) {
    return "mysql";
  }
  return "postgresql";
}

// Convert PostgreSQL parameter placeholders ($1, $2) to MySQL (?)
function convertParams(sql: string, dbType: DatabaseType): string {
  if (dbType === "mysql") {
    // Replace $1, $2, etc. with ?
    return sql.replace(/\$\d+/g, "?");
  }
  return sql;
}

class PostgreSQLPool implements DatabasePool {
  private pool: PgPool;

  constructor(connectionString: string) {
    this.pool = new PgPool({ connectionString });
  }

  async query(sql: string, params?: unknown[]) {
    const result = await this.pool.query(sql, params);
    return { rows: result.rows };
  }

  async end() {
    await this.pool.end();
  }
}

class MySQLPool implements DatabasePool {
  private pool: mysql.Pool;

  constructor(connectionString: string) {
    // Parse mysql:// URL to connection config
    const url = new URL(connectionString);
    this.pool = mysql.createPool({
      host: url.hostname,
      port: parseInt(url.port || "3306"),
      user: url.username,
      password: url.password,
      database: url.pathname.slice(1), // Remove leading /
      waitForConnections: true,
      connectionLimit: 10,
    });
  }

  async query(sql: string, params?: unknown[]) {
    // MySQL uses ? placeholders, convert from PostgreSQL $1, $2 style
    const convertedSql = convertParams(sql, "mysql");
    const [rows] = await this.pool.execute(convertedSql, params);
    return { rows: rows as Record<string, unknown>[] };
  }

  async end() {
    await this.pool.end();
  }
}

export function createPool(connectionString: string): DatabasePool {
  const dbType = detectDatabaseType(connectionString);

  if (dbType === "mysql") {
    return new MySQLPool(connectionString);
  }

  return new PostgreSQLPool(connectionString);
}
