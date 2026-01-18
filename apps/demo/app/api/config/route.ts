import { NextResponse } from "next/server";
import { readFile } from "fs/promises";
import path from "path";
import { detectDatabaseType } from "@/lib/db-factory";

export async function GET() {
  try {
    const connectionString = process.env.PRIMARY_DATABASE_URL || "";
    const dbType = detectDatabaseType(connectionString);

    const filename = dbType === "mysql" ? "transforms-mysql.yml" : "transforms-pg.yml";
    const filePath = path.join(process.cwd(), "public", filename);
    const content = await readFile(filePath, "utf-8");

    const source =
      dbType === "mysql"
        ? "environments/mysql-development/config/transforms.yml"
        : "environments/pg-development/config/transforms.yml";

    return NextResponse.json({ content, source });
  } catch (err) {
    console.error("Config file error:", err);
    return NextResponse.json({ error: "Failed to read config file", details: String(err) }, { status: 500 });
  }
}
