package main

import (
	"context"
	"fmt"
	"os"

	"github.com/betting-platform/internal/core/usecase"
	"github.com/betting-platform/internal/core/usecase/games"
	"github.com/betting-platform/internal/core/usecase/tax"
	"github.com/betting-platform/internal/core/usecase/wallet"
	"github.com/betting-platform/internal/infrastructure/config"
	"github.com/betting-platform/internal/infrastructure/database"
	"github.com/betting-platform/internal/infrastructure/events"
	gameshttp "github.com/betting-platform/internal/infrastructure/http"
	"github.com/betting-platform/internal/infrastructure/http/health"
	"github.com/betting-platform/internal/infrastructure/http/middleware"
	"github.com/betting-platform/internal/infrastructure/logging"
	"github.com/betting-platform/internal/infrastructure/metrics"
	"github.com/betting-platform/internal/infrastructure/ratelimit"
	"github.com/betting-platform/internal/infrastructure/repository/postgres"
	"github.com/betting-platform/internal/infrastructure/server"
	"github.com/betting-platform/internal/infrastructure/websocket"
	"github.com/gorilla/mux"
)

// redisRateLimiterChecker implements health.Checker for Redis rate limiter
type redisRateLimiterChecker struct {
	limiter *ratelimit.RedisLimiter
}

func (r *redisRateLimiterChecker) Name() string {
	return "redis_rate_limiter"
}

func (r *redisRateLimiterChecker) Check(ctx context.Context) error {
	return r.limiter.HealthCheck(ctx)
}

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}
	logger := logging.Setup(cfg.Logging.Level, cfg.Logging.Format)
	logger.Info("starting games", "env", cfg.Service.Environment)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	hub := websocket.NewHub()
	go hub.Run()

	fairService := usecase.NewProvablyFairService()

	db, err := database.NewPostgresConnection(database.Config{
		Host:            cfg.Database.Host,
		Port:            cfg.Database.Port,
		User:            cfg.Database.User,
		Password:        cfg.Database.Password,
		DBName:          cfg.Database.Name,
		SSLMode:         cfg.Database.SSLMode,
		MaxOpenConns:    cfg.Database.MaxOpenConns,
		MaxIdleConns:    cfg.Database.MaxIdleConns,
		ConnMaxLifetime: cfg.Database.ConnMaxLifetime,
	})
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	gameRepo := postgres.NewGameRepository(db)
	betRepo := postgres.NewGameBetRepository(db)

	// Initialize wallet service and tax engine
	walletService := wallet.New(db)
	taxEngine, err := tax.Default()
	if err != nil {
		logger.Error("failed to create tax engine", "error", err)
		os.Exit(1)
	}

	// Initialize event bus
	eventBus, err := events.Connect(cfg.NATS.URL, "games-service")
	if err != nil {
		logger.Error("failed to connect to NATS", "error", err)
		os.Exit(1)
	}
	defer eventBus.Close()

	// Initialize MaxMind provider (optional)
	geoProvider, err := middleware.NewMaxMindProvider(middleware.DefaultDBPath())
	if err != nil {
		logger.Warn("failed to load maxmind database", "error", err)
	}

	engine := games.NewCrashGameEngine(hub, fairService, gameRepo, betRepo, walletService, taxEngine)
	go engine.StartGame(ctx)

	r := mux.NewRouter()

	// Initialize metrics
	rec := metrics.New("games")
	rec.RegisterRoutes(r)

	// Apply middleware stack
	r.Use(middleware.RequestID)
	r.Use(middleware.Recovery)
	r.Use(middleware.Logging)
	r.Use(middleware.CORS(cfg.Security))
	r.Use(rec.Middleware)
	r.Use(middleware.Geolocation(middleware.GeoConfig{
		Provider: geoProvider,
		Allowed:  cfg.Tenant.AllowedCountries,
	}))

	// Use Redis-backed rate limiter to prevent memory leak DDoS vulnerability
	rateLimitConfig := &ratelimit.Config{
		IPRequestsPerWindow:     cfg.Security.RateLimitRequests,
		IPWindow:                cfg.Security.RateLimitWindow,
		UserRequestsPerWindow:   cfg.Security.RateLimitRequests * 2, // Allow more for authenticated users
		UserWindow:              cfg.Security.RateLimitWindow,
		GlobalRequestsPerWindow: cfg.Security.RateLimitRequests * 10, // Global limit
		GlobalWindow:            cfg.Security.RateLimitWindow,
		RedisAddr:               cfg.Redis.Addr(),
		RedisPassword:           cfg.Redis.Password,
		RedisDB:                 0,
		UserPrefix:              "rate_limit:games:user:",
		IPPrefix:                "rate_limit:games:ip:",
		GlobalPrefix:            "rate_limit:games:global:",
	}

	// Use existing Redis rate limiter directly
	redisRateLimiter, err := ratelimit.NewRedisLimiter(ctx, rateLimitConfig)
	if err != nil {
		logger.Error("failed to create Redis rate limiter", "error", err)
		os.Exit(1)
	}
	defer redisRateLimiter.Close()

	// Create proxy validator for X-Forwarded-For spoofing prevention
	proxyValidator, err := middleware.NewDefaultProxyValidator([]string{}) // Add trusted proxy IPs if needed
	if err != nil {
		logger.Error("failed to create proxy validator", "error", err)
		os.Exit(1)
	}

	// Apply rate limiting middleware
	r.Use(middleware.RateLimitMiddleware(redisRateLimiter, rateLimitConfig, proxyValidator))

	h := health.NewHandler("games", "dev")
	h.Register(&health.PostgresChecker{DB: db})

	// Register Redis rate limiter health check
	h.Register(&redisRateLimiterChecker{limiter: redisRateLimiter})

	h.RegisterRoutes(r)

	gameshttp.NewGamesHandler(engine).RegisterRoutes(r)

	addr := fmt.Sprintf(":%d", cfg.Service.Port)
	if err := server.RunHTTP(ctx, addr, r, logger); err != nil {
		logger.Error("server terminated with error", "error", err)
		os.Exit(1)
	}
}
