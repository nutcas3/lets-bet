-- Drop sports betting tables in reverse order of creation

-- Drop triggers first
DROP TRIGGER IF EXISTS update_sport_bets_updated_at ON sport_bets;
DROP TRIGGER IF EXISTS update_market_outcomes_updated_at ON market_outcomes;
DROP TRIGGER IF EXISTS update_betting_markets_updated_at ON betting_markets;
DROP TRIGGER IF EXISTS update_sport_events_updated_at ON sport_events;

-- Drop functions
DROP FUNCTION IF EXISTS update_sport_bets_updated_at();
DROP FUNCTION IF EXISTS update_market_outcomes_updated_at();
DROP FUNCTION IF EXISTS update_betting_markets_updated_at();
DROP FUNCTION IF EXISTS update_sport_events_updated_at();

-- Drop tables
DROP TABLE IF EXISTS odds_history;
DROP TABLE IF EXISTS sport_bets;
DROP TABLE IF EXISTS market_outcomes;
DROP TABLE IF EXISTS betting_markets;
DROP TABLE IF EXISTS sport_events;
