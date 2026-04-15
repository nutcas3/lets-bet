package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
	"github.com/shopspring/decimal"
)

// Config holds database connection configuration
type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// NewPostgresConnection creates a new PostgreSQL connection
func NewPostgresConnection(cfg Config) (*sql.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Successfully connected to PostgreSQL database")
	return db, nil
}

// NullDecimal provides a nullable decimal type for database operations
type NullDecimal struct {
	decimal.Decimal
	Valid bool
}

// Scan implements the sql.Scanner interface
func (nd *NullDecimal) Scan(value interface{}) error {
	if value == nil {
		nd.Valid = false
		return nil
	}

	switch v := value.(type) {
	case []byte:
		d, err := decimal.NewFromString(string(v))
		if err != nil {
			return err
		}
		nd.Decimal = d
		nd.Valid = true
	case string:
		d, err := decimal.NewFromString(v)
		if err != nil {
			return err
		}
		nd.Decimal = d
		nd.Valid = true
	default:
		return fmt.Errorf("cannot scan %T into NullDecimal", value)
	}

	return nil
}

// Value implements the driver.Valuer interface
func (nd NullDecimal) Value() (interface{}, error) {
	if !nd.Valid {
		return nil, nil
	}
	return nd.Decimal.String(), nil
}

// NewNullDecimal creates a new NullDecimal
func NewNullDecimal(d decimal.Decimal) NullDecimal {
	return NullDecimal{
		Decimal: d,
		Valid:   true,
	}
}

// GetDefaultConfig returns default database configuration
func GetDefaultConfig() Config {
	return Config{
		Host:     "localhost",
		Port:     "5432",
		User:     "betting_user",
		Password: "betting_password",
		DBName:   "betting_platform",
		SSLMode:  "disable",
	}
}
