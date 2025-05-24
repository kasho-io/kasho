import { NextResponse } from 'next/server';
import { queryReplica } from '@/lib/db';

export async function GET() {
  try {
    const data = await queryReplica(`
      SELECT id, organization_id, name, email, password, created_at, updated_at
      FROM users
      LIMIT 5
    `);
    return NextResponse.json(data);
  } catch (err) {
    return NextResponse.json({ error: 'Failed to fetch data', details: String(err) }, { status: 500 });
  }
} 