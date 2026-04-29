-- Create sports betting tables

-- Create sport_events table for storing sports events from odds feeds
CREATE TABLE IF NOT EXISTS sport_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id VARCHAR(255) UNIQUE NOT NULL, -- External event ID from odds provider
    sport VARCHAR(50) NOT NULL,
    tournament VARCHAR(255) NOT NULL,
    home_team VARCHAR(255) NOT NULL,
    away_team VARCHAR(255) NOT NULL,
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'SCHEDULED' CHECK (status IN ('SCHEDULED', 'LIVE', 'FINISHED', 'POSTPONED', 'CANCELLED')),
    
    -- Score information
    home_score INTEGER DEFAULT 0,
    away_score INTEGER DEFAULT 0,
    home_score_half_time INTEGER,
    away_score_half_time INTEGER,
    
    -- Provider information
    provider VARCHAR(50) NOT NULL DEFAULT 'sportradar', -- 'sportradar', 'genius'
    provider_data JSONB DEFAULT '{}',
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Constraints
    CONSTRAINT sport_events_status_check CHECK (status IN ('SCHEDULED', 'LIVE', 'FINISHED', 'POSTPONED', 'CANCELLED'))
);

-- Create betting markets table
CREATE TABLE IF NOT EXISTS betting_markets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    market_id VARCHAR(255) UNIQUE NOT NULL, -- External market ID
    event_id UUID NOT NULL REFERENCES sport_events(id) ON DELETE CASCADE,
    market_type VARCHAR(50) NOT NULL, -- 'MATCH_WINNER', 'HANDICAP', 'TOTAL_GOALS', etc.
    market_name VARCHAR(255) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'OPEN' CHECK (status IN ('OPEN', 'SUSPENDED', 'CLOSED', 'SETTLED')),
    suspended_at TIMESTAMP WITH TIME ZONE,
    provider VARCHAR(50) NOT NULL DEFAULT 'sportradar',
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create market outcomes table
CREATE TABLE IF NOT EXISTS market_outcomes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    outcome_id VARCHAR(255) UNIQUE NOT NULL, -- External outcome ID
    market_id UUID NOT NULL REFERENCES betting_markets(id) ON DELETE CASCADE,
    outcome_name VARCHAR(255) NOT NULL,
    odds DECIMAL(10,4) NOT NULL CHECK (odds > 0),
    price DECIMAL(10,4) NOT NULL CHECK (price > 0),
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING' CHECK (status IN ('PENDING', 'WON', 'LOST', 'VOID')),
    
    -- Settlement information
    settled_at TIMESTAMP WITH TIME ZONE,
    settlement_factor DECIMAL(10,4) DEFAULT 1.0, -- For partial settlements
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create sport_bets table for user bets
CREATE TABLE IF NOT EXISTS sport_bets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    bet_id VARCHAR(255) UNIQUE NOT NULL, -- Internal bet ID
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    event_id UUID NOT NULL REFERENCES sport_events(id) ON DELETE CASCADE,
    market_id UUID NOT NULL REFERENCES betting_markets(id) ON DELETE CASCADE,
    outcome_id UUID NOT NULL REFERENCES market_outcomes(id) ON DELETE CASCADE,
    
    -- Bet details
    amount DECIMAL(15,2) NOT NULL CHECK (amount > 0),
    odds DECIMAL(10,4) NOT NULL CHECK (odds > 0),
    currency VARCHAR(3) NOT NULL DEFAULT 'KES',
    
    -- Bet status
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING' CHECK (status IN ('PENDING', 'WON', 'LOST', 'VOID', 'CANCELLED')),
    payout DECIMAL(15,2) DEFAULT 0,
    net_payout DECIMAL(15,2) DEFAULT 0, -- After taxes
    
    -- Settlement information
    settled_at TIMESTAMP WITH TIME ZONE,
    settlement_reason TEXT,
    
    -- Timestamps
    placed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Constraints
    CONSTRAINT sport_bets_status_check CHECK (status IN ('PENDING', 'WON', 'LOST', 'VOID', 'CANCELLED'))
);

-- Create odds history table for tracking odds changes
CREATE TABLE IF NOT EXISTS odds_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    outcome_id UUID NOT NULL REFERENCES market_outcomes(id) ON DELETE CASCADE,
    old_odds DECIMAL(10,4),
    new_odds DECIMAL(10,4) NOT NULL,
    old_price DECIMAL(10,4),
    new_price DECIMAL(10,4) NOT NULL,
    change_reason VARCHAR(100),
    provider VARCHAR(50) NOT NULL DEFAULT 'sportradar',
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for sport_events
CREATE INDEX IF NOT EXISTS idx_sport_events_event_id ON sport_events(event_id);
CREATE INDEX IF NOT EXISTS idx_sport_events_sport ON sport_events(sport);
CREATE INDEX IF NOT EXISTS idx_sport_events_tournament ON sport_events(tournament);
CREATE INDEX IF NOT EXISTS idx_sport_events_status ON sport_events(status);
CREATE INDEX IF NOT EXISTS idx_sport_events_start_time ON sport_events(start_time);
CREATE INDEX IF NOT EXISTS idx_sport_events_provider ON sport_events(provider);
CREATE INDEX IF NOT EXISTS idx_sport_events_created_at ON sport_events(created_at);

-- Create indexes for betting_markets
CREATE INDEX IF NOT EXISTS idx_betting_markets_market_id ON betting_markets(market_id);
CREATE INDEX IF NOT EXISTS idx_betting_markets_event_id ON betting_markets(event_id);
CREATE INDEX IF NOT EXISTS idx_betting_markets_market_type ON betting_markets(market_type);
CREATE INDEX IF NOT EXISTS idx_betting_markets_status ON betting_markets(status);
CREATE INDEX IF NOT EXISTS idx_betting_markets_provider ON betting_markets(provider);

-- Create indexes for market_outcomes
CREATE INDEX IF NOT EXISTS idx_market_outcomes_outcome_id ON market_outcomes(outcome_id);
CREATE INDEX IF NOT EXISTS idx_market_outcomes_market_id ON market_outcomes(market_id);
CREATE INDEX IF NOT EXISTS idx_market_outcomes_status ON market_outcomes(status);
CREATE INDEX IF NOT EXISTS idx_market_outcomes_odds ON market_outcomes(odds);

-- Create indexes for sport_bets
CREATE INDEX IF NOT EXISTS idx_sport_bets_bet_id ON sport_bets(bet_id);
CREATE INDEX IF NOT EXISTS idx_sport_bets_user_id ON sport_bets(user_id);
CREATE INDEX IF NOT EXISTS idx_sport_bets_event_id ON sport_bets(event_id);
CREATE INDEX IF NOT EXISTS idx_sport_bets_market_id ON sport_bets(market_id);
CREATE INDEX IF NOT EXISTS idx_sport_bets_outcome_id ON sport_bets(outcome_id);
CREATE INDEX IF NOT EXISTS idx_sport_bets_status ON sport_bets(status);
CREATE INDEX IF NOT EXISTS idx_sport_bets_placed_at ON sport_bets(placed_at);
CREATE INDEX IF NOT EXISTS idx_sport_bets_amount ON sport_bets(amount);

-- Create indexes for odds_history
CREATE INDEX IF NOT EXISTS idx_odds_history_outcome_id ON odds_history(outcome_id);
CREATE INDEX IF NOT EXISTS idx_odds_history_created_at ON odds_history(created_at);

-- Create updated_at triggers
CREATE OR REPLACE FUNCTION update_sport_events_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_sport_events_updated_at
    BEFORE UPDATE ON sport_events
    FOR EACH ROW
    EXECUTE FUNCTION update_sport_events_updated_at();

CREATE OR REPLACE FUNCTION update_betting_markets_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_betting_markets_updated_at
    BEFORE UPDATE ON betting_markets
    FOR EACH ROW
    EXECUTE FUNCTION update_betting_markets_updated_at();

CREATE OR REPLACE FUNCTION update_market_outcomes_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_market_outcomes_updated_at
    BEFORE UPDATE ON market_outcomes
    FOR EACH ROW
    EXECUTE FUNCTION update_market_outcomes_updated_at();

CREATE OR REPLACE FUNCTION update_sport_bets_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_sport_bets_updated_at
    BEFORE UPDATE ON sport_bets
    FOR EACH ROW
    EXECUTE FUNCTION update_sport_bets_updated_at();

-- Add comments
COMMENT ON TABLE sport_events IS 'Stores sports events from odds feed providers';
COMMENT ON TABLE betting_markets IS 'Stores betting markets for sports events';
COMMENT ON TABLE market_outcomes IS 'Stores outcomes for betting markets with odds';
COMMENT ON TABLE sport_bets IS 'Stores user bets on sports events';
COMMENT ON TABLE odds_history IS 'Tracks odds changes over time';

-- Add column comments for sport_events
COMMENT ON COLUMN sport_events.id IS 'Primary key';
COMMENT ON COLUMN sport_events.event_id IS 'External event ID from odds provider';
COMMENT ON COLUMN sport_events.sport IS 'Sport type (football, basketball, etc.)';
COMMENT ON COLUMN sport_events.tournament IS 'Tournament/league name';
COMMENT ON COLUMN sport_events.home_team IS 'Home team name';
COMMENT ON COLUMN sport_events.away_team IS 'Away team name';
COMMENT ON COLUMN sport_events.start_time IS 'Event start time';
COMMENT ON COLUMN sport_events.status IS 'Event status';
COMMENT ON COLUMN sport_events.home_score IS 'Home team score';
COMMENT ON COLUMN sport_events.away_score IS 'Away team score';
COMMENT ON COLUMN sport_events.provider IS 'Odds provider name';
COMMENT ON COLUMN sport_events.provider_data IS 'Raw provider data in JSON';

-- Add column comments for betting_markets
COMMENT ON COLUMN betting_markets.id IS 'Primary key';
COMMENT ON COLUMN betting_markets.market_id IS 'External market ID';
COMMENT ON COLUMN betting_markets.event_id IS 'Reference to sport event';
COMMENT ON COLUMN betting_markets.market_type IS 'Type of betting market';
COMMENT ON COLUMN betting_markets.market_name IS 'Display name of market';
COMMENT ON COLUMN betting_markets.status IS 'Market status';
COMMENT ON COLUMN betting_markets.suspended_at IS 'When market was suspended';

-- Add column comments for market_outcomes
COMMENT ON COLUMN market_outcomes.id IS 'Primary key';
COMMENT ON COLUMN market_outcomes.outcome_id IS 'External outcome ID';
COMMENT ON COLUMN market_outcomes.market_id IS 'Reference to betting market';
COMMENT ON COLUMN market_outcomes.outcome_name IS 'Display name of outcome';
COMMENT ON COLUMN market_outcomes.odds IS 'Decimal odds';
COMMENT ON COLUMN market_outcomes.price IS 'Alternative odds format';
COMMENT ON COLUMN market_outcomes.status IS 'Outcome status';
COMMENT ON COLUMN market_outcomes.settled_at IS 'When outcome was settled';
COMMENT ON COLUMN market_outcomes.settlement_factor IS 'Factor for partial settlements';

-- Add column comments for sport_bets
COMMENT ON COLUMN sport_bets.id IS 'Primary key';
COMMENT ON COLUMN sport_bets.bet_id IS 'Internal bet identifier';
COMMENT ON COLUMN sport_bets.user_id IS 'User who placed the bet';
COMMENT ON COLUMN sport_bets.event_id IS 'Sport event reference';
COMMENT ON COLUMN sport_bets.market_id IS 'Betting market reference';
COMMENT ON COLUMN sport_bets.outcome_id IS 'Outcome reference';
COMMENT ON COLUMN sport_bets.amount IS 'Bet amount';
COMMENT ON COLUMN sport_bets.odds IS 'Odds at time of bet';
COMMENT ON COLUMN sport_bets.currency IS 'Bet currency';
COMMENT ON COLUMN sport_bets.status IS 'Bet status';
COMMENT ON COLUMN sport_bets.payout IS 'Gross payout amount';
COMMENT ON COLUMN sport_bets.net_payout IS 'Net payout after taxes';
COMMENT ON COLUMN sport_bets.settled_at IS 'When bet was settled';
COMMENT ON COLUMN sport_bets.settlement_reason IS 'Reason for settlement';
COMMENT ON COLUMN sport_bets.placed_at IS 'When bet was placed';

-- Add column comments for odds_history
COMMENT ON COLUMN odds_history.id IS 'Primary key';
COMMENT ON COLUMN odds_history.outcome_id IS 'Reference to market outcome';
COMMENT ON COLUMN odds_history.old_odds IS 'Previous odds value';
COMMENT ON COLUMN odds_history.new_odds IS 'New odds value';
COMMENT ON COLUMN odds_history.old_price IS 'Previous price value';
COMMENT ON COLUMN odds_history.new_price IS 'New price value';
COMMENT ON COLUMN odds_history.change_reason IS 'Reason for odds change';
COMMENT ON COLUMN odds_history.provider IS 'Odds provider';
COMMENT ON COLUMN odds_history.created_at IS 'When odds change occurred';
