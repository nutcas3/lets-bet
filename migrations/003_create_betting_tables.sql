-- Create betting-related tables
-- Supports sports betting and crash games

CREATE TABLE IF NOT EXISTS bets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    country_code VARCHAR(2) NOT NULL,
    
    -- Bet details
    bet_type VARCHAR(10) NOT NULL, -- SINGLE, MULTI, SYSTEM
    stake DECIMAL(15,2) NOT NULL,
    currency VARCHAR(3) NOT NULL,
    potential_win DECIMAL(15,2) NOT NULL,
    total_odds DECIMAL(10,3) NOT NULL,
    
    -- Status and settlement
    status VARCHAR(20) DEFAULT 'PENDING', -- PENDING, WON, LOST, VOID, CASHED_OUT
    actual_win DECIMAL(15,2) DEFAULT 0.00,
    settled_at TIMESTAMP,
    
    -- Metadata
    placed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    ip_address INET,
    device_id VARCHAR(100),
    
    -- Tax tracking (BCLB compliance)
    tax_amount DECIMAL(15,2) DEFAULT 0.00,
    tax_paid BOOLEAN DEFAULT FALSE
);

-- Create selections table for bet details
CREATE TABLE IF NOT EXISTS bet_selections (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    bet_id UUID NOT NULL REFERENCES bets(id) ON DELETE CASCADE,
    market_id VARCHAR(100) NOT NULL, -- From odds provider
    event_id VARCHAR(100) NOT NULL, -- e.g., "match_12345"
    event_name VARCHAR(255) NOT NULL, -- e.g., "Arsenal vs Chelsea"
    market_name VARCHAR(100) NOT NULL, -- e.g., "Match Winner"
    outcome_name VARCHAR(100) NOT NULL, -- e.g., "Arsenal"
    odds DECIMAL(10,3) NOT NULL,
    status VARCHAR(20) DEFAULT 'PENDING', -- PENDING, WON, LOST, VOID
    settled_at TIMESTAMP
);

-- Create games table for crash games
CREATE TABLE IF NOT EXISTS games (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    game_type VARCHAR(20) NOT NULL, -- CRASH, VIRTUAL, SLOT
    round_number BIGINT NOT NULL,
    
    -- Provably Fair seeds
    server_seed VARCHAR(255) NOT NULL,
    server_seed_hash VARCHAR(255) NOT NULL,
    client_seed VARCHAR(255) NOT NULL,
    
    -- Game result
    crash_point DECIMAL(10,3) NOT NULL, -- e.g., 2.45x
    
    -- State
    status VARCHAR(20) DEFAULT 'WAITING', -- WAITING, RUNNING, CRASHED
    started_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    crashed_at TIMESTAMP,
    
    -- Metadata
    country_code VARCHAR(2) NOT NULL,
    min_bet DECIMAL(15,2) DEFAULT 10.00,
    max_bet DECIMAL(15,2) DEFAULT 10000.00,
    max_multiplier DECIMAL(10,3) DEFAULT 100.00
);

-- Create game_bets table for crash game betting
CREATE TABLE IF NOT EXISTS game_bets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    game_id UUID NOT NULL REFERENCES games(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    
    amount DECIMAL(15,2) NOT NULL,
    currency VARCHAR(3) NOT NULL,
    
    -- Cashout
    cashed_out BOOLEAN DEFAULT FALSE,
    cashout_at DECIMAL(10,3), -- Multiplier when cashed out
    payout DECIMAL(15,2) DEFAULT 0.00,
    
    status VARCHAR(20) DEFAULT 'ACTIVE', -- ACTIVE, WON, LOST, CASHED_OUT
    placed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    cashed_out_at TIMESTAMP
);

-- Create matches table for sports betting
CREATE TABLE IF NOT EXISTS matches (
    id VARCHAR(100) PRIMARY KEY, -- External ID from odds provider
    sport VARCHAR(20) NOT NULL, -- FOOTBALL, BASKETBALL, TENNIS, CRICKET, RUGBY
    league VARCHAR(100) NOT NULL,
    home_team VARCHAR(255) NOT NULL,
    away_team VARCHAR(255) NOT NULL,
    start_time TIMESTAMP NOT NULL,
    status VARCHAR(20) DEFAULT 'SCHEDULED', -- SCHEDULED, LIVE, FINISHED, POSTPONED, CANCELLED
    country_code VARCHAR(2) NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create match_markets table
CREATE TABLE IF NOT EXISTS match_markets (
    id VARCHAR(100) PRIMARY KEY, -- External ID from odds provider
    match_id VARCHAR(100) NOT NULL REFERENCES matches(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL, -- MATCH_WINNER, HANDICAP, TOTAL_GOALS, etc.
    name VARCHAR(100) NOT NULL,
    status VARCHAR(20) DEFAULT 'OPEN', -- OPEN, SUSPENDED, CLOSED, SETTLED
    suspended_at TIMESTAMP
);

-- Create match_outcomes table
CREATE TABLE IF NOT EXISTS match_outcomes (
    id VARCHAR(100) PRIMARY KEY, -- External ID from odds provider
    market_id VARCHAR(100) NOT NULL REFERENCES match_markets(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    odds DECIMAL(10,3) NOT NULL,
    price DECIMAL(10,3), -- Alternative odds format
    status VARCHAR(20) DEFAULT 'PENDING' -- PENDING, WON, LOST, VOID
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_bets_user_id ON bets(user_id);
CREATE INDEX IF NOT EXISTS idx_bets_status ON bets(status);
CREATE INDEX IF NOT EXISTS idx_bets_placed_at ON bets(placed_at);
CREATE INDEX IF NOT EXISTS idx_bet_selections_bet_id ON bet_selections(bet_id);
CREATE INDEX IF NOT EXISTS idx_bet_selections_event_id ON bet_selections(event_id);
CREATE INDEX IF NOT EXISTS idx_games_status ON games(status);
CREATE INDEX IF NOT EXISTS idx_games_round_number ON games(round_number);
CREATE INDEX IF NOT EXISTS idx_game_bets_game_id ON game_bets(game_id);
CREATE INDEX IF NOT EXISTS idx_game_bets_user_id ON game_bets(user_id);
CREATE INDEX IF NOT EXISTS idx_matches_status ON matches(status);
CREATE INDEX IF NOT EXISTS idx_matches_start_time ON matches(start_time);
CREATE INDEX IF NOT EXISTS idx_match_markets_match_id ON match_markets(match_id);
CREATE INDEX IF NOT EXISTS idx_match_outcomes_market_id ON match_outcomes(market_id);
