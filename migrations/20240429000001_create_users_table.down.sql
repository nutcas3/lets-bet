-- Drop users table

-- Drop trigger
DROP TRIGGER IF EXISTS update_users_updated_at ON users;

-- Drop function (only if no other tables use it)
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop table
DROP TABLE IF EXISTS users;
