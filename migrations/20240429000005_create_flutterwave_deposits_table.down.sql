-- Drop flutterwave deposits table

-- Drop trigger
DROP TRIGGER IF EXISTS update_flutterwave_deposits_updated_at ON flutterwave_deposits;

-- Drop function
DROP FUNCTION IF EXISTS update_flutterwave_deposits_updated_at();

-- Drop table
DROP TABLE IF EXISTS flutterwave_deposits;
