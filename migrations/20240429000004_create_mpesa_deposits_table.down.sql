-- Drop mpesa deposits table

-- Drop trigger
DROP TRIGGER IF EXISTS update_mpesa_deposits_updated_at ON mpesa_deposits;

-- Drop table
DROP TABLE IF EXISTS mpesa_deposits;
