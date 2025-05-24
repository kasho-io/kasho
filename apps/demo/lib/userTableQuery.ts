export const USER_TABLE_QUERY = `
  SELECT id, organization_id, name, email, password, created_at, updated_at
  FROM users
  LIMIT 8
`; 