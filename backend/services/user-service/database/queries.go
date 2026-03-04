package database

// SQL queries for user repository.
// Centralized here so schema/table names can be configured in one place.

const InsertUserQuery = `
INSERT INTO service.users (
  id,
  user_name,
  full_name,
  phone,
  password_hash,
  salt,
  status,
  is_verified,
  user_id,
  provider,
  provider_user_id,
  access_token,
  password_changed_at,
  refresh_token,
  created_by,
  updated_by,
  created_at,
  updated_at,
  expires_at
) VALUES (
  ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
)
`

const UpdateUserQuery = `
UPDATE service.users
SET
  full_name = ?,
  user_name = ?,
  access_token = ?,
  updated_at = ?
WHERE id = ?
`
