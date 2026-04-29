package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/betting-platform/internal/infrastructure/config"
	"github.com/betting-platform/internal/infrastructure/logging"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	dir := flag.String("dir", "./migrations", "migrations directory")
	action := flag.String("action", "up", "migration action: up, down, version")
	steps := flag.Int("steps", 0, "number of steps to migrate (0 = all)")
	flag.Parse()

	cfg, err := config.LoadConfig()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	logger := logging.Setup(cfg.Logging.Level, cfg.Logging.Format)

	// Build database URL
	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Name,
		cfg.Database.SSLMode,
	)

	// Create migrate instance
	m, err := migrate.New(
		"file://"+*dir,
		dbURL,
	)
	if err != nil {
		logger.Error("failed to create migrate instance", "error", err)
		os.Exit(1)
	}
	defer m.Close()

	// Execute migration action
	switch *action {
	case "up":
		if *steps > 0 {
			if err := m.Steps(*steps); err != nil && err != migrate.ErrNoChange {
				logger.Error("migration failed", "error", err)
				os.Exit(1)
			}
		} else {
			if err := m.Up(); err != nil && err != migrate.ErrNoChange {
				logger.Error("migration failed", "error", err)
				os.Exit(1)
			}
		}
		logger.Info("migrations applied successfully")
	case "down":
		if *steps > 0 {
			if err := m.Steps(-*steps); err != nil && err != migrate.ErrNoChange {
				logger.Error("migration failed", "error", err)
				os.Exit(1)
			}
		} else {
			if err := m.Down(); err != nil && err != migrate.ErrNoChange {
				logger.Error("migration failed", "error", err)
				os.Exit(1)
			}
		}
		logger.Info("migrations rolled back successfully")
	case "version":
		version, dirty, err := m.Version()
		if err != nil {
			logger.Error("failed to get version", "error", err)
			os.Exit(1)
		}
		logger.Info("current migration version", "version", version, "dirty", dirty)
	default:
		logger.Error("invalid action", "action", *action)
		os.Exit(1)
	}
}
