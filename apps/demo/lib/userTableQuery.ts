export const USER_TABLE_QUERY = `
  SELECT id, organization_id, name, email, password, updated_at
  FROM users
  ORDER BY id ASC
  LIMIT 5
`; 