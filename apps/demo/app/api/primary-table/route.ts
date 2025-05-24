import { NextResponse } from 'next/server';
import { queryPrimary } from '@/lib/db';
import { USER_TABLE_QUERY } from '@/lib/userTableQuery';

export async function GET() {
  try {
    const data = await queryPrimary(USER_TABLE_QUERY);
    return NextResponse.json(data);
  } catch (err) {
    console.error('DB error:', err);
    return NextResponse.json({ error: 'Failed to fetch data', details: String(err) }, { status: 500 });
  }
}

export async function PATCH(req: Request) {
  try {
    const { rows } = await req.json();
    if (!Array.isArray(rows)) {
      return NextResponse.json({ error: 'Invalid payload' }, { status: 400 });
    }
    for (const row of rows) {
      if (!row.id) continue;
      await queryPrimary(
        'UPDATE users SET name = $1, email = $2 WHERE id = $3',
        [row.name, row.email, row.id]
      );
    }
    return NextResponse.json({ success: true });
  } catch (err) {
    console.error('PATCH error:', err);
    return NextResponse.json({ error: 'Failed to update data', details: String(err) }, { status: 500 });
  }
} 