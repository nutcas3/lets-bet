-- Drop wallets and transactions tables

-- Drop trigger
DROP TRIGGER IF EXISTS update_wallets_updated_at ON wallets;

-- Drop tables
DROP TABLE IF EXISTS transactions;
DROP TABLE IF EXISTS wallets;
