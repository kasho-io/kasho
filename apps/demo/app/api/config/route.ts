import { NextResponse } from "next/server";
import { readFile } from "fs/promises";
import path from "path";

export async function GET() {
  try {
    // Read the transforms.yml file from the public directory (symlinked)
    const filePath = path.join(process.cwd(), "public", "transforms.yml");
    const content = await readFile(filePath, "utf-8");

    return NextResponse.json({
      content,
      source: "environments/demo/config/transforms.yml",
    });
  } catch (err) {
    console.error("Config file error:", err);
    return NextResponse.json({ error: "Failed to read config file", details: String(err) }, { status: 500 });
  }
}
