package postgres

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/betting-platform/internal/core/domain"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

// UserRepository implements user repository using PostgreSQL
type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (
			id, phone_number, email, password_hash, country_code, currency,
			national_id, kra_pin, full_name, date_of_birth, is_verified,
			self_excluded, self_excluded_until, daily_deposit_limit,
			created_at, updated_at, status
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
	`

	_, err := r.db.ExecContext(ctx, query,
		user.ID, user.PhoneNumber, user.Email, user.PasswordHash, user.CountryCode, user.Currency,
		user.NationalID, user.KRAPin, user.FullName, user.DateOfBirth, user.IsVerified,
		user.SelfExcluded, user.SelfExcludedUntil, user.DailyDepositLimit,
		user.CreatedAt, user.UpdatedAt, user.Status,
	)

	if err != nil {
		log.Printf("Error creating user: %v", err)
		return err
	}

	return nil
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	query := `
		SELECT id, phone_number, email, password_hash, country_code, currency,
			   national_id, kra_pin, full_name, date_of_birth, is_verified,
			   self_excluded, self_excluded_until, daily_deposit_limit,
			   created_at, updated_at, last_login_at, status
		FROM users WHERE id = $1
	`

	var user domain.User
	var selfExcludedUntil sql.NullTime
	var lastLoginAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID, &user.PhoneNumber, &user.Email, &user.PasswordHash, &user.CountryCode, &user.Currency,
		&user.NationalID, &user.KRAPin, &user.FullName, &user.DateOfBirth, &user.IsVerified,
		&user.SelfExcluded, &selfExcludedUntil, &user.DailyDepositLimit,
		&user.CreatedAt, &user.UpdatedAt, &lastLoginAt, &user.Status,
	)

	if err != nil {
		return nil, err
	}

	if selfExcludedUntil.Valid {
		user.SelfExcludedUntil = &selfExcludedUntil.Time
	}

	if lastLoginAt.Valid {
		user.LastLoginAt = &lastLoginAt.Time
	}

	return &user, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT id, phone_number, email, password_hash, country_code, currency,
			   national_id, kra_pin, full_name, date_of_birth, is_verified,
			   self_excluded, self_excluded_until, daily_deposit_limit,
			   created_at, updated_at, last_login_at, status
		FROM users WHERE email = $1
	`

	var user domain.User
	var selfExcludedUntil sql.NullTime
	var lastLoginAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID, &user.PhoneNumber, &user.Email, &user.PasswordHash, &user.CountryCode, &user.Currency,
		&user.NationalID, &user.KRAPin, &user.FullName, &user.DateOfBirth, &user.IsVerified,
		&user.SelfExcluded, &selfExcludedUntil, &user.DailyDepositLimit,
		&user.CreatedAt, &user.UpdatedAt, &lastLoginAt, &user.Status,
	)

	if err != nil {
		return nil, err
	}

	if selfExcludedUntil.Valid {
		user.SelfExcludedUntil = &selfExcludedUntil.Time
	}

	if lastLoginAt.Valid {
		user.LastLoginAt = &lastLoginAt.Time
	}

	return &user, nil
}

func (r *UserRepository) Update(ctx context.Context, user *domain.User) error {
	query := `
		UPDATE users SET 
			phone_number = $2, email = $3, password_hash = $4, country_code = $5, currency = $6,
			national_id = $7, kra_pin = $8, full_name = $9, date_of_birth = $10, is_verified = $11,
			self_excluded = $12, self_excluded_until = $13, daily_deposit_limit = $14,
			updated_at = $15, status = $16
		WHERE id = $1
	`

	now := time.Now()

	_, err := r.db.ExecContext(ctx, query,
		user.ID, user.PhoneNumber, user.Email, user.PasswordHash, user.CountryCode, user.Currency,
		user.NationalID, user.KRAPin, user.FullName, user.DateOfBirth, user.IsVerified,
		user.SelfExcluded, user.SelfExcludedUntil, user.DailyDepositLimit,
		now, user.Status,
	)

	if err != nil {
		log.Printf("Error updating user: %v", err)
		return err
	}

	return nil
}

func (r *UserRepository) UpdateVerificationStatus(ctx context.Context, id uuid.UUID, isVerified bool) error {
	query := `UPDATE users SET is_verified = $2, updated_at = $3 WHERE id = $1`

	now := time.Now()

	_, err := r.db.ExecContext(ctx, query, id, isVerified, now)
	if err != nil {
		log.Printf("Error updating verification status: %v", err)
		return err
	}

	return nil
}
