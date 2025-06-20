import { NextResponse } from 'next/server';
import { createHash } from 'crypto';
import * as argon2 from 'argon2';
import { queryPrimary } from '@/lib/db';
import { USER_TABLE_QUERY } from '@/lib/userTableQuery';

// Replicate Go's generateDeterministicSalt function
function generateDeterministicSalt(original: string, length: number): Buffer {
  const hash = createHash('sha256');
  hash.update(original);
  const fullHash = hash.digest();
  
  // If we need more bytes than SHA256 provides, cycle through the hash
  const salt = Buffer.alloc(length);
  for (let i = 0; i < length; i++) {
    salt[i] = fullHash[i % fullHash.length];
  }
  return salt;
}

// Replicate Go's TransformPasswordArgon2id function exactly
async function transformPasswordArgon2id(cleartext: string, useSalt: boolean, time: number, memory: number, threads: number, original: string): Promise<string> {
  let salt: Buffer;
  if (useSalt) {
    salt = generateDeterministicSalt(original, 16); // 16 bytes salt
  } else {
    salt = Buffer.alloc(16); // Empty salt
  }
  
  // Generate hash using raw argon2id - matches Go's argon2.IDKey
  const hash = await argon2.hash(cleartext, {
    type: argon2.argon2id,
    memoryCost: memory,
    timeCost: time,
    parallelism: threads,
    salt: salt,
    hashLength: 32, // 32 bytes output
    raw: true, // Get raw bytes instead of encoded string
  });
  
  // Format: salt$hash (both hex encoded) - matches Go format exactly
  return `${salt.toString('hex')}$${(hash as Buffer).toString('hex')}`;
}

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
      
      // Build dynamic update query based on provided fields
      const updates = [];
      const values = [];
      let paramIndex = 1;
      
      if (row.name !== undefined) {
        updates.push(`name = $${paramIndex++}`);
        values.push(row.name);
      }
      if (row.email !== undefined) {
        updates.push(`email = $${paramIndex++}`);
        values.push(row.email);
      }
      if (row.password !== undefined) {
        // Hash the cleartext password with Argon2id before storing
        // Use exact same parameters as Go implementation
        const hashedPassword = await transformPasswordArgon2id(
          row.password, // cleartext
          true, // useSalt
          3, // time (default)
          65536, // memory (default: 64MB in KiB)
          4, // threads (default)
          row.password // original (for deterministic salt)
        );
        updates.push(`password = $${paramIndex++}`);
        values.push(hashedPassword);
      }
      
      if (updates.length > 0) {
        // Always update the updated_at timestamp
        updates.push(`updated_at = NOW()`);
        values.push(row.id); // Add ID as the last parameter
        await queryPrimary(
          `UPDATE users SET ${updates.join(', ')} WHERE id = $${paramIndex}`,
          values
        );
      }
    }
    return NextResponse.json({ success: true });
  } catch (err) {
    console.error('PATCH error:', err);
    return NextResponse.json({ error: 'Failed to update data', details: String(err) }, { status: 500 });
  }
} 