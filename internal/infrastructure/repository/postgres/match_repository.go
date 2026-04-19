package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/betting-platform/internal/core/domain"
)

// MatchRepository implements match repository using PostgreSQL
type MatchRepository struct {
	db *sql.DB
}

// NewMatchRepository creates a new match repository
func NewMatchRepository(db *sql.DB) *MatchRepository {
	return &MatchRepository{db: db}
}

// Create creates a new match
func (r *MatchRepository) Create(ctx context.Context, match *domain.Match) error {
	query := `
		INSERT INTO sport_events (
			id, event_id, sport, tournament, home_team, away_team,
			start_time, status, home_score, away_score,
			provider, provider_data, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`

	_, err := r.db.ExecContext(ctx, query,
		match.ID, match.ID, string(match.Sport), match.League, match.HomeTeam, match.AwayTeam,
		match.StartTime, string(match.Status), match.Score.HomeScore, match.Score.AwayScore,
		"system", "{}", time.Now(), time.Now(),
	)

	return err
}

// GetByID retrieves a match by ID
func (r *MatchRepository) GetByID(ctx context.Context, id string) (*domain.Match, error) {
	query := `
		SELECT id, sport, tournament, home_team, away_team, start_time, status,
			   home_score, away_score, created_at, updated_at
		FROM sport_events
		WHERE id = $1
	`

	var match domain.Match
	var sport, league, homeTeam, awayTeam string
	var startTime time.Time
	var status domain.MatchStatus
	var countryCode string
	var createdAt, updatedAt time.Time
	var homeScore, awayScore int

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&match.ID, &sport, &league, &homeTeam, &awayTeam,
		&startTime, &status, &homeScore, &awayScore, &createdAt, &updatedAt,
	)

	if err != nil {
		return nil, err
	}

	match.Sport = domain.Sport(sport)
	match.League = league
	match.Status = domain.MatchStatus(status)

	if homeScore > 0 || awayScore > 0 {
		match.Score = &domain.MatchScore{
			HomeScore: homeScore,
			AwayScore: awayScore,
		}
	}

	match.StartTime = startTime
	match.CountryCode = countryCode

	return &match, nil
}

// GetByEventID retrieves a match by event ID
func (r *MatchRepository) GetByEventID(ctx context.Context, eventID string) (*domain.Match, error) {
	query := `
		SELECT id, sport, tournament, home_team, away_team, start_time, status,
			   home_score, away_score, created_at, updated_at
		FROM sport_events
		WHERE event_id = $1
	`

	var match domain.Match
	var homeScore, awayScore sql.NullInt32
	var sport, tournament, status string
	var createdAt, updatedAt time.Time

	err := r.db.QueryRowContext(ctx, query, eventID).Scan(
		&match.ID, &sport, &tournament, &match.HomeTeam, &match.AwayTeam,
		&match.StartTime, &status, &homeScore, &awayScore,
		&createdAt, &updatedAt,
	)

	if err != nil {
		return nil, err
	}

	match.Sport = domain.Sport(sport)
	match.League = tournament
	match.Status = domain.MatchStatus(status)

	if homeScore.Valid && awayScore.Valid {
		match.Score = &domain.MatchScore{
			HomeScore: int(homeScore.Int32),
			AwayScore: int(awayScore.Int32),
		}
	}

	return &match, nil
}

// Update updates a match
func (r *MatchRepository) Update(ctx context.Context, match *domain.Match) error {
	query := `
		UPDATE sport_events
		SET sport = $2, tournament = $3, home_team = $4, away_team = $5,
			start_time = $6, status = $7, home_score = $8, away_score = $9,
			updated_at = $10
		WHERE id = $1
	`

	homeScore := sql.NullInt32{Int32: 0, Valid: false}
	awayScore := sql.NullInt32{Int32: 0, Valid: false}

	if match.Score != nil {
		homeScore = sql.NullInt32{Int32: int32(match.Score.HomeScore), Valid: true}
		awayScore = sql.NullInt32{Int32: int32(match.Score.AwayScore), Valid: true}
	}

	_, err := r.db.ExecContext(ctx, query,
		match.ID, string(match.Sport), match.League, match.HomeTeam, match.AwayTeam,
		match.StartTime, string(match.Status), homeScore, awayScore, time.Now(),
	)

	return err
}

// UpdateStatus updates match status
func (r *MatchRepository) UpdateStatus(ctx context.Context, id string, status domain.MatchStatus) error {
	query := `
		UPDATE sport_events
		SET status = $2, updated_at = $3
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query, id, string(status), time.Now())
	return err
}

// UpdateScore updates match score
func (r *MatchRepository) UpdateScore(ctx context.Context, id string, score *domain.MatchScore) error {
	query := `
		UPDATE sport_events
		SET home_score = $2, away_score = $3, updated_at = $4
		WHERE id = $1
	`

	homeScore := sql.NullInt32{Int32: 0, Valid: false}
	awayScore := sql.NullInt32{Int32: 0, Valid: false}

	if score != nil {
		homeScore = sql.NullInt32{Int32: int32(score.HomeScore), Valid: true}
		awayScore = sql.NullInt32{Int32: int32(score.AwayScore), Valid: true}
	}

	_, err := r.db.ExecContext(ctx, query, id, homeScore, awayScore, time.Now())
	return err
}

// GetLiveMatches retrieves all live matches
func (r *MatchRepository) GetLiveMatches(ctx context.Context) ([]*domain.Match, error) {
	query := `
		SELECT id, sport, tournament, home_team, away_team, start_time, status,
			   home_score, away_score, created_at, updated_at
		FROM sport_events
		WHERE status = 'LIVE'
		ORDER BY start_time ASC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var matches []*domain.Match
	for rows.Next() {
		var match domain.Match
		var homeScore, awayScore sql.NullInt32
		var sport, tournament, status string
		var createdAt, updatedAt time.Time

		err := rows.Scan(
			&match.ID, &sport, &tournament, &match.HomeTeam, &match.AwayTeam,
			&match.StartTime, &status, &homeScore, &awayScore,
			&createdAt, &updatedAt,
		)

		if err != nil {
			return nil, err
		}

		match.Sport = domain.Sport(sport)
		match.League = tournament
		match.Status = domain.MatchStatus(status)

		if homeScore.Valid && awayScore.Valid {
			match.Score = &domain.MatchScore{
				HomeScore: int(homeScore.Int32),
				AwayScore: int(awayScore.Int32),
			}
		}

		matches = append(matches, &match)
	}

	return matches, nil
}

// GetBySport retrieves matches by sport
func (r *MatchRepository) GetBySport(ctx context.Context, sport domain.Sport) ([]*domain.Match, error) {
	query := `
		SELECT id, sport, tournament, home_team, away_team, start_time, status,
			   home_score, away_score, created_at, updated_at
		FROM sport_events
		WHERE sport = $1
		ORDER BY start_time ASC
	`

	rows, err := r.db.QueryContext(ctx, query, string(sport))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var matches []*domain.Match
	for rows.Next() {
		var match domain.Match
		var homeScore, awayScore sql.NullInt32
		var sportStr, tournament, status string
		var createdAt, updatedAt time.Time

		err := rows.Scan(
			&match.ID, &sportStr, &tournament, &match.HomeTeam, &match.AwayTeam,
			&match.StartTime, &status, &homeScore, &awayScore,
			&createdAt, &updatedAt,
		)

		if err != nil {
			return nil, err
		}

		match.Sport = domain.Sport(sportStr)
		match.League = tournament
		match.Status = domain.MatchStatus(status)

		if homeScore.Valid && awayScore.Valid {
			match.Score = &domain.MatchScore{
				HomeScore: int(homeScore.Int32),
				AwayScore: int(awayScore.Int32),
			}
		}

		matches = append(matches, &match)
	}

	return matches, nil
}

// Delete deletes a match
func (r *MatchRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM sport_events WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}
