-- Enable extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    phone_number VARCHAR(20) NOT NULL,
    email VARCHAR(255),
    password_hash VARCHAR(255) NOT NULL,
    country_code VARCHAR(2) NOT NULL, -- KE, NG, GH
    currency VARCHAR(3) NOT NULL, -- KES, NGN, GHS
    
    -- KYC (BCLB Compliance)
    national_id VARCHAR(50),
    kra_pin VARCHAR(20),
    full_name VARCHAR(255) NOT NULL,
    date_of_birth DATE NOT NULL,
    is_verified BOOLEAN DEFAULT FALSE,
    
    -- Responsible Gaming
    self_excluded BOOLEAN DEFAULT FALSE,
    self_excluded_until TIMESTAMP,
    daily_deposit_limit BIGINT, -- in cents
    
    -- Metadata
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    last_login_at TIMESTAMP,
    status VARCHAR(20) DEFAULT 'ACTIVE',
    
    CONSTRAINT unique_phone_country UNIQUE (phone_number, country_code)
);

CREATE INDEX idx_users_country ON users(country_code);
CREATE INDEX idx_users_phone ON users(phone_number);
CREATE INDEX idx_users_status ON users(status);

CREATE TABLE wallets (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id),
    currency VARCHAR(3) NOT NULL,
    balance DECIMAL(20, 2) DEFAULT 0.00,
    version BIGINT DEFAULT 0, -- For optimistic locking
    bonus_balance DECIMAL(20, 2) DEFAULT 0.00,
    
    -- Daily limits tracking
    today_deposit DECIMAL(20, 2) DEFAULT 0.00,
    last_deposit_reset TIMESTAMP DEFAULT NOW(),
    
    updated_at TIMESTAMP DEFAULT NOW(),
    
    CONSTRAINT unique_user_wallet UNIQUE (user_id, currency)
);

CREATE INDEX idx_wallets_user ON wallets(user_id);

CREATE TABLE transactions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    wallet_id UUID NOT NULL REFERENCES wallets(id),
    user_id UUID NOT NULL REFERENCES users(id),
    
    type VARCHAR(20) NOT NULL, -- DEPOSIT, WITHDRAWAL, BET_PLACED, BET_WON, etc.
    amount DECIMAL(20, 2) NOT NULL,
    currency VARCHAR(3) NOT NULL,
    
    -- Balance snapshot (for auditing)
    balance_before DECIMAL(20, 2) NOT NULL,
    balance_after DECIMAL(20, 2) NOT NULL,
    
    -- Reference
    reference_id UUID,
    reference_type VARCHAR(20), -- BET, GAME, etc.
    
    -- Payment provider (for deposits/withdrawals)
    provider_txn_id VARCHAR(255),
    provider_name VARCHAR(50), -- MPESA, AIRTEL, PAYSTACK
    
    status VARCHAR(20) DEFAULT 'PENDING',
    description TEXT,
    
    created_at TIMESTAMP DEFAULT NOW(),
    completed_at TIMESTAMP,
    
    -- Tenant
    country_code VARCHAR(2) NOT NULL
);

CREATE INDEX idx_transactions_user ON transactions(user_id);
CREATE INDEX idx_transactions_wallet ON transactions(wallet_id);
CREATE INDEX idx_transactions_type ON transactions(type);
CREATE INDEX idx_transactions_status ON transactions(status);
CREATE INDEX idx_transactions_country ON transactions(country_code);
CREATE INDEX idx_transactions_reference ON transactions(reference_id, reference_type);
CREATE INDEX idx_transactions_created ON transactions(created_at DESC);

-- ============================================
-- BETS TABLE
-- ============================================
CREATE TABLE bets (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id),
    country_code VARCHAR(2) NOT NULL,
    
    bet_type VARCHAR(20) NOT NULL, -- SINGLE, MULTI, SYSTEM
    stake DECIMAL(20, 2) NOT NULL,
    currency VARCHAR(3) NOT NULL,
    potential_win DECIMAL(20, 2) NOT NULL,
    total_odds DECIMAL(10, 2) NOT NULL,
    
    -- Settlement
    status VARCHAR(20) DEFAULT 'PENDING',
    actual_win DECIMAL(20, 2) DEFAULT 0.00,
    settled_at TIMESTAMP,
    
    -- Metadata
    placed_at TIMESTAMP DEFAULT NOW(),
    ip_address INET,
    device_id VARCHAR(255),
    
    -- Tax (BCLB)
    tax_amount DECIMAL(20, 2) DEFAULT 0.00,
    tax_paid BOOLEAN DEFAULT FALSE
);

CREATE INDEX idx_bets_user ON bets(user_id);
CREATE INDEX idx_bets_status ON bets(status);
CREATE INDEX idx_bets_country ON bets(country_code);
CREATE INDEX idx_bets_placed ON bets(placed_at DESC);

CREATE TABLE bet_selections (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    bet_id UUID NOT NULL REFERENCES bets(id) ON DELETE CASCADE,
    
    market_id VARCHAR(255) NOT NULL,
    event_id VARCHAR(255) NOT NULL,
    event_name VARCHAR(500),
    market_name VARCHAR(255),
    outcome_name VARCHAR(255),
    odds DECIMAL(10, 2) NOT NULL,
    
    status VARCHAR(20) DEFAULT 'PENDING',
    settled_at TIMESTAMP
);

CREATE INDEX idx_selections_bet ON bet_selections(bet_id);
CREATE INDEX idx_selections_event ON bet_selections(event_id);
CREATE INDEX idx_selections_status ON bet_selections(status);

CREATE TABLE games (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    game_type VARCHAR(20) NOT NULL, -- CRASH, VIRTUAL, SLOT
    round_number BIGINT NOT NULL,
    
    -- Provably Fair
    server_seed VARCHAR(255) NOT NULL,
    server_seed_hash VARCHAR(255) NOT NULL,
    client_seed VARCHAR(255) NOT NULL,
    
    -- Result
    crash_point DECIMAL(10, 2),
    
    -- State
    status VARCHAR(20) DEFAULT 'WAITING',
    started_at TIMESTAMP DEFAULT NOW(),
    crashed_at TIMESTAMP,
    
    -- Config
    country_code VARCHAR(2),
    min_bet DECIMAL(20, 2) DEFAULT 10.00,
    max_bet DECIMAL(20, 2) DEFAULT 10000.00,
    max_multiplier DECIMAL(10, 2) DEFAULT 100.00
);

CREATE INDEX idx_games_type ON games(game_type);
CREATE INDEX idx_games_status ON games(status);
CREATE INDEX idx_games_round ON games(round_number DESC);
CREATE INDEX idx_games_started ON games(started_at DESC);

CREATE TABLE game_bets (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    game_id UUID NOT NULL REFERENCES games(id),
    user_id UUID NOT NULL REFERENCES users(id),
    
    amount DECIMAL(20, 2) NOT NULL,
    currency VARCHAR(3) NOT NULL,
    
    -- Cashout
    cashed_out BOOLEAN DEFAULT FALSE,
    cashout_at DECIMAL(10, 2), -- Multiplier at cashout
    payout DECIMAL(20, 2) DEFAULT 0.00,
    
    status VARCHAR(20) DEFAULT 'ACTIVE',
    placed_at TIMESTAMP DEFAULT NOW(),
    cashed_out_at TIMESTAMP
);

CREATE INDEX idx_game_bets_game ON game_bets(game_id);
CREATE INDEX idx_game_bets_user ON game_bets(user_id);
CREATE INDEX idx_game_bets_status ON game_bets(status);

CREATE TABLE audit_log (
    id BIGSERIAL PRIMARY KEY,
    user_id UUID,
    action VARCHAR(50) NOT NULL,
    entity_type VARCHAR(50),
    entity_id UUID,
    old_value JSONB,
    new_value JSONB,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    country_code VARCHAR(2)
);

CREATE INDEX idx_audit_user ON audit_log(user_id);
CREATE INDEX idx_audit_action ON audit_log(action);
CREATE INDEX idx_audit_created ON audit_log(created_at DESC);

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_wallets_updated_at
    BEFORE UPDATE ON wallets
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
