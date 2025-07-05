import { NextResponse } from "next/server";
import { queryReplica } from "@/lib/db";
import { USER_TABLE_QUERY } from "@/lib/userTableQuery";

export async function GET() {
  try {
    const data = await queryReplica(USER_TABLE_QUERY);
    return NextResponse.json(data);
  } catch (err) {
    console.error("DB error:", err);
    return NextResponse.json({ error: "Failed to fetch data", details: String(err) }, { status: 500 });
  }
}
