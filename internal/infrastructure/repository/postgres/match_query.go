package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/betting-platform/internal/core/domain"
)

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

	return r.scanMatches(rows)
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

	return r.scanMatches(rows)
}

// scanMatches scans multiple matches from rows
func (r *MatchRepository) scanMatches(rows *sql.Rows) ([]*domain.Match, error) {
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
